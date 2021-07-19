// Copyright 2020 PingCAP, Inc. Licensed under Apache-2.0.

package restore

import (
	"context"
	"sync"

	"github.com/pingcap/br/pkg/metautil"
	"github.com/pingcap/br/pkg/utils"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/pingcap/parser/model"
	"go.uber.org/zap"

	"github.com/pingcap/br/pkg/glue"
	"github.com/pingcap/br/pkg/rtree"
)

const (
	defaultChannelSize = 1024
)

// TableSink is the 'sink' of restored data by a sender.
type TableSink interface {
	EmitTables(tables ...CreatedTable)
	EmitError(error)
	Close()
}

type chanTableSink struct {
	outCh chan<- []CreatedTable
	errCh chan<- error
}

func (sink chanTableSink) EmitTables(tables ...CreatedTable) {
	sink.outCh <- tables
}

func (sink chanTableSink) EmitError(err error) {
	sink.errCh <- err
}

func (sink chanTableSink) Close() {
	// ErrCh may has multi sender part, don't close it.
	close(sink.outCh)
}

// ContextManager is the struct to manage a TiKV 'context' for restore.
// Batcher will call Enter when any table should be restore on batch,
// so you can do some prepare work here(e.g. set placement rules for online restore).
type ContextManager interface {
	// Enter make some tables 'enter' this context(a.k.a., prepare for restore).
	Enter(ctx context.Context, tables []CreatedTable) error
	// Leave make some tables 'leave' this context(a.k.a., restore is done, do some post-works).
	Leave(ctx context.Context, tables []CreatedTable) error
	// Close closes the context manager, sometimes when the manager is 'killed' and should do some cleanup
	// it would be call.
	Close(ctx context.Context)
}

// NewBRContextManager makes a BR context manager, that is,
// set placement rules for online restore when enter(see <splitPrepareWork>),
// unset them when leave.
func NewBRContextManager(client *Client) ContextManager {
	return &brContextManager{
		client: client,

		hasTable: make(map[int64]CreatedTable),
	}
}

type brContextManager struct {
	client *Client

	// This 'set' of table ID allow us to handle each table just once.
	hasTable map[int64]CreatedTable
}

func (manager *brContextManager) Close(ctx context.Context) {
	tbls := make([]*model.TableInfo, 0, len(manager.hasTable))
	for _, tbl := range manager.hasTable {
		tbls = append(tbls, tbl.Table)
	}
	splitPostWork(ctx, manager.client, tbls)
}

func (manager *brContextManager) Enter(ctx context.Context, tables []CreatedTable) error {
	placementRuleTables := make([]*model.TableInfo, 0, len(tables))

	for _, tbl := range tables {
		if _, ok := manager.hasTable[tbl.Table.ID]; !ok {
			placementRuleTables = append(placementRuleTables, tbl.Table)
		}
		manager.hasTable[tbl.Table.ID] = tbl
	}

	return splitPrepareWork(ctx, manager.client, placementRuleTables)
}

func (manager *brContextManager) Leave(ctx context.Context, tables []CreatedTable) error {
	placementRuleTables := make([]*model.TableInfo, 0, len(tables))

	for _, table := range tables {
		placementRuleTables = append(placementRuleTables, table.Table)
	}

	splitPostWork(ctx, manager.client, placementRuleTables)
	log.Info("restore table done", ZapTables(tables))
	for _, tbl := range placementRuleTables {
		delete(manager.hasTable, tbl.ID)
	}
	return nil
}

func splitPostWork(ctx context.Context, client *Client, tables []*model.TableInfo) {
	err := client.ResetPlacementRules(ctx, tables)
	if err != nil {
		log.Warn("reset placement rules failed", zap.Error(err))
		return
	}
}

func splitPrepareWork(ctx context.Context, client *Client, tables []*model.TableInfo) error {
	err := client.SetupPlacementRules(ctx, tables)
	if err != nil {
		log.Error("setup placement rules failed", zap.Error(err))
		return errors.Trace(err)
	}

	err = client.WaitPlacementSchedule(ctx, tables)
	if err != nil {
		log.Error("wait placement schedule failed", zap.Error(err))
		return errors.Trace(err)
	}
	return nil
}

// CreatedTable is a table created on restore process,
// but not yet filled with data.
type CreatedTable struct {
	RewriteRule *RewriteRules
	Table       *model.TableInfo
	OldTable    *metautil.Table
}

// TableWithRange is a CreatedTable that has been bind to some of key ranges.
type TableWithRange struct {
	CreatedTable

	Range []rtree.Range
}

// Exhaust drains all remaining errors in the channel, into a slice of errors.
func Exhaust(ec <-chan error) []error {
	out := make([]error, 0, len(ec))
	for {
		select {
		case err := <-ec:
			out = append(out, err)
		default:
			// errCh will NEVER be closed(ya see, it has multi sender-part),
			// so we just consume the current backlog of this channel, then return.
			return out
		}
	}
}

// BatchSender is the abstract of how the batcher send a batch.
type BatchSender interface {
	// PutSink sets the sink of this sender, user to this interface promise
	// call this function at least once before first call to `RestoreBatch`.
	PutSink(sink TableSink)
	// RestoreBatch will send the restore request.
	RestoreBatch(ranges DrainResult)
	Close()
}

type tikvSender struct {
	client   *Client
	updateCh glue.Progress

	sink TableSink
	inCh chan<- DrainResult

	wg *sync.WaitGroup

	tableWaiters *sync.Map
}

func (b *tikvSender) PutSink(sink TableSink) {
	// don't worry about visibility, since we will call this before first call to
	// RestoreBatch, which is a sync point.
	b.sink = sink
}

func (b *tikvSender) RestoreBatch(ranges DrainResult) {
	log.Info("restore batch: waiting ranges", zap.Int("range", len(b.inCh)))
	b.inCh <- ranges
}

// NewTiKVSender make a sender that send restore requests to TiKV.
func NewTiKVSender(
	ctx context.Context,
	cli *Client,
	updateCh glue.Progress,
	splitConcurrency uint,
) (BatchSender, error) {
	inCh := make(chan DrainResult, defaultChannelSize)
	midCh := make(chan drainResultAndDone, defaultChannelSize)

	sender := &tikvSender{
		client:       cli,
		updateCh:     updateCh,
		inCh:         inCh,
		wg:           new(sync.WaitGroup),
		tableWaiters: new(sync.Map),
	}

	sender.wg.Add(2)
	go sender.splitWorker(ctx, inCh, midCh, splitConcurrency)
	go sender.restoreWorker(ctx, midCh)
	return sender, nil
}

func (b *tikvSender) Close() {
	close(b.inCh)
	b.wg.Wait()
	log.Debug("tikv sender closed")
}

type drainResultAndDone struct {
	result DrainResult
	done   func()
}

func (b *tikvSender) splitWorker(ctx context.Context,
	ranges <-chan DrainResult,
	next chan<- drainResultAndDone,
	concurrency uint,
) {
	defer log.Debug("split worker closed")
	splitWorks := new(sync.WaitGroup)
	defer func() {
		splitWorks.Wait()
		b.wg.Done()
		close(next)
	}()
	pool := utils.NewWorkerPool(concurrency, "split")
	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-ranges:
			if !ok {
				return
			}
			splitWorks.Add(1)
			done := b.registerTableIsRestoring(result.TablesToSend)
			pool.Apply(func() {
				SplitRangesAndThen(ctx, b.client, result.Ranges, result.RewriteRules, b.updateCh, func(err error) {
					if err != nil {
						log.Error("failed on split range", rtree.ZapRanges(result.Ranges), zap.Error(err))
						b.sink.EmitError(err)
						return
					}
					next <- drainResultAndDone{
						result: result,
						done:   done,
					}
					splitWorks.Done()
				})
			})
		}
	}
}

func (b *tikvSender) registerTableIsRestoring(ts []CreatedTable) func() {
	wgs := make([]*sync.WaitGroup, 0, len(ts))
	for _, t := range ts {
		i, _ := b.tableWaiters.LoadOrStore(t.Table.ID, new(sync.WaitGroup))
		wg := i.(*sync.WaitGroup)
		wg.Add(1)
		wgs = append(wgs, wg)
	}
	return func() {
		for _, wg := range wgs {
			wg.Done()
		}
	}
}

func (b *tikvSender) waitTablesDone(ts []CreatedTable) {
	for _, t := range ts {
		wg, ok := b.tableWaiters.LoadAndDelete(t.Table.ID)
		if !ok {
			log.Panic("bug! table done before register!",
				zap.Any("wait-table-map", b.tableWaiters),
				zap.Stringer("table", t.Table.Name))
		}
		wg.(*sync.WaitGroup).Wait()
	}
}

func (b *tikvSender) restoreWorker(ctx context.Context, ranges <-chan drainResultAndDone) {
	restoreWorks := new(sync.WaitGroup)
	defer func() {
		log.Debug("restore worker closed")
		restoreWorks.Wait()
		b.wg.Done()
		b.sink.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-ranges:
			if !ok {
				return
			}
			restoreWorks.Add(1)
			files := r.result.Files()
			// There has been a worker in the `RestoreFiles` procedure.
			// Spawning a raw goroutine won't make too many requests to TiKV.
			go b.client.RestoreFilesAndThen(ctx, files, r.result.RewriteRules, b.updateCh, func(e error) {
				if e != nil {
					b.sink.EmitError(e)
				}
				log.Info("restore batch done", rtree.ZapRanges(r.result.Ranges))
				r.done()
				b.waitTablesDone(r.result.BlankTablesAfterSend)
				b.sink.EmitTables(r.result.BlankTablesAfterSend...)
				restoreWorks.Done()
			})
		}
	}
}
