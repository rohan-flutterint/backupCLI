// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package local

import (
	"bytes"
	"context"
	"io"
	"sort"
	"sync"
	"time"

	split "github.com/pingcap/br/pkg/restore"

	"github.com/docker/go-units"
	"github.com/pingcap/kvproto/pkg/kvrpcpb"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/parser/model"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/tidb/distsql"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/table"
	"github.com/pingcap/tidb/util/codec"
	"github.com/pingcap/tidb/util/ranger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	backendkv "github.com/pingcap/br/pkg/lightning/backend/kv"
	"github.com/pingcap/br/pkg/lightning/common"
	"github.com/pingcap/br/pkg/lightning/log"
	"github.com/pingcap/br/pkg/logutil"

	"github.com/cockroachdb/pebble"
	"github.com/pingcap/errors"
	sst "github.com/pingcap/kvproto/pkg/import_sstpb"
	kvrpc "github.com/pingcap/kvproto/pkg/kvrpcpb"
	tikv "github.com/pingcap/kvproto/pkg/tikvpb"
)

const (
	maxWriteBatchCount        = 128
	defaultEngineMemCacheSize = 512 * units.MiB
	maxScanRegionSize         = 256
)

type DuplicateRequest struct {
	tableID   int64
	indexID   int64 // 0 represent it is a table request
	start     kv.Key
	end       kv.Key
	indexInfo *model.IndexInfo
}

type DuplicateManager struct {
	db                *pebble.DB
	splitCli          split.SplitClient
	regionConcurrency int
	connPool          common.GRPCConns
	tls               *common.TLS
	sqlMode           mysql.SQLMode
	ts                uint64
}

func NewDuplicateManager(
	db *pebble.DB,
	splitCli split.SplitClient,
	ts uint64,
	tls *common.TLS,
	regionConcurrency int,
	sqlMode mysql.SQLMode) (*DuplicateManager, error) {
	return &DuplicateManager{
		db:                db,
		tls:               tls,
		regionConcurrency: regionConcurrency,
		sqlMode:           sqlMode,
		splitCli:          splitCli,
		ts:                ts,
	}, nil
}

func (manager *DuplicateManager) DuplicateTable(ctx context.Context, tbl table.Table) error {
	reqs, err := buildDuplicateRequests(tbl.Meta())
	if err != nil {
		return err
	}

	decoder, err := backendkv.NewTableKVDecoder(tbl, &backendkv.SessionOptions{
		SQLMode: manager.sqlMode,
	})
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var tableErr common.OnceError
	rpcctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, r := range reqs {
		wg.Add(1)
		go func(req *DuplicateRequest) {
			err := manager.sendRequestToTiKV(rpcctx, decoder, req)
			if err != nil {
				log.L().Error("error occur when collect duplicate data from TiKV", zap.Error(err))
				tableErr.Set(err)
				cancel()
			}
			wg.Done()
		}(r)
	}
	wg.Wait()
	return nil
}

func (manager *DuplicateManager) sendRequestToTiKV(ctx context.Context, decoder *backendkv.TableKVDecoder, req *DuplicateRequest) error {
	startKey := codec.EncodeBytes([]byte{}, req.start)
	endKey := codec.EncodeBytes([]byte{}, req.end)

	regions, err := paginateScanRegion(ctx, manager.splitCli, startKey, endKey, 1024)
	if err != nil {
		return err
	}
	tryTimes := 0
	indexHandles := make([][]byte, len(regions))
	for {
		if len(regions) == 0 {
			break
		}
		if tryTimes > maxRetryTimes {
			return errors.Errorf("retry time exceed limit")
		}
		unfinishedRegions := make([]*split.RegionInfo, len(regions))
		waitingClients := make([]sst.ImportSST_DuplicateDetectClient, len(regions))
		watingRegions := make([]*split.RegionInfo, len(regions))
		for idx, region := range regions {
			if len(waitingClients) > manager.regionConcurrency {
				r := regions[idx:]
				unfinishedRegions = append(unfinishedRegions, r...)
				break
			}
			_, start, _ := codec.DecodeBytes(region.Region.StartKey, []byte{})
			_, end, _ := codec.DecodeBytes(region.Region.EndKey, []byte{})
			if bytes.Compare(startKey, region.Region.StartKey) > 0 {
				start = req.start
			}
			if region.Region.EndKey == nil || len(region.Region.EndKey) == 0 || bytes.Compare(endKey, region.Region.EndKey) < 0 {
				end = req.end
			}

			cli, err := manager.getDuplicateStream(ctx, region, start, end)
			if err != nil {
				r, err := manager.splitCli.GetRegionByID(ctx, region.Region.GetId())
				if err != nil {
					unfinishedRegions = append(unfinishedRegions, region)
				} else {
					unfinishedRegions = append(unfinishedRegions, r)
				}
			} else {
				waitingClients = append(waitingClients, cli)
				watingRegions = append(watingRegions, region)
			}
		}

		if len(indexHandles) > 0 {
			handles := manager.GetValues(ctx, indexHandles)
			if len(handles) > 0 {
				indexHandles = handles
			} else {
				indexHandles = indexHandles[:0]
			}
		}

		for idx, cli := range waitingClients {
			region := watingRegions[idx]
			for {
				resp, reqErr := cli.Recv()
				hasErr := false
				if reqErr != nil {
					if errors.Cause(reqErr) == io.EOF {
						break
					} else {
						hasErr = true
					}
				}

				if hasErr || resp.GetKeyError() != nil {
					r, err := manager.splitCli.GetRegionByID(ctx, region.Region.GetId())
					if err != nil {
						unfinishedRegions = append(unfinishedRegions, region)
					} else {
						unfinishedRegions = append(unfinishedRegions, r)
					}
				}
				if hasErr {
					log.L().Warn("meet error when recving duplicate detect response from TiKV, retry again",
						logutil.Region(region.Region), logutil.Leader(region.Leader), zap.Error(reqErr))
					break
				}
				if resp.GetKeyError() != nil {
					log.L().Warn("meet key error in duplicate detect response from TiKV, retry again ",
						logutil.Region(region.Region), logutil.Leader(region.Leader),
						zap.String("KeyError", resp.GetKeyError().GetMessage()))
					break
				}

				if resp.GetRegionError() != nil {
					log.L().Warn("meet key error in duplicate detect response from TiKV, retry again ",
						logutil.Region(region.Region), logutil.Leader(region.Leader),
						zap.String("RegionError", resp.GetRegionError().GetMessage()))

					r, err := paginateScanRegion(ctx, manager.splitCli, watingRegions[idx].Region.GetStartKey(), watingRegions[idx].Region.GetEndKey(), 1024)
					if err != nil {
						unfinishedRegions = append(unfinishedRegions, watingRegions[idx])
					} else {
						unfinishedRegions = append(unfinishedRegions, r...)
					}
					break
				}

				handles, err := manager.storeDuplicateData(ctx, region, resp, decoder, req)
				if err != nil {
					return err
				}
				if handles != nil && len(handles) > 0 {
					indexHandles = append(indexHandles, handles...)
				}
			}
		}

		// it means that all of region send to TiKV fail, so we must sleep some time to avoid retry too frequency
		if len(unfinishedRegions) == len(regions) {
			tryTimes += 1
			time.Sleep(defaultRetryBackoffTime)
		}
		regions = unfinishedRegions
	}
	return nil
}

func (manager *DuplicateManager) storeDuplicateData(
	ctx context.Context,
	region *split.RegionInfo,
	resp *sst.DuplicateDetectResponse, decoder *backendkv.TableKVDecoder, req *DuplicateRequest,
) ([][]byte, error) {
	opts := &pebble.WriteOptions{Sync: false}
	var err error
	for i := 0; i < maxRetryTimes; i++ {
		b := manager.db.NewBatch()
		handles := make([][]byte, len(resp.Pairs))
		for _, kv := range resp.Pairs {
			if req.indexInfo != nil {
				h, err := decoder.DecodeHandleFromIndex(req.indexInfo, kv.Key, kv.Value)
				if err != nil {
					log.L().Error("decode handle error from index",
						zap.Error(err), logutil.Key("key", kv.Key),
						logutil.Key("value", kv.Value), zap.Uint64("commit-ts", kv.CommitTs))
					continue
				}
				key := decoder.EncodeHandleKey(h)
				handles = append(handles, key)
			} else {
				b.Set(kv.Key, kv.Value, opts)
			}
		}
		err = b.Commit(opts)
		if err != nil {
			continue
		}
		err := manager.getValuesFromRegion(ctx, region, handles)
		if err == nil {
			return nil, nil
		} else {
			// Retry kv get handles after
			log.L().Error("failed to collect values from TiKV by handle, we will retry it", zap.Error(err))
			return handles, nil
		}
	}
	return nil, err
}

func (manager *DuplicateManager) ReportDuplicateData() error {
	// TODO
	return nil
}

func (manager *DuplicateManager) RepairDuplicateData() error {
	// TODO
	return nil
}

func (manager *DuplicateManager) GetValues(
	ctx context.Context,
	handles [][]byte,
) [][]byte {
	retryHandles := make([][]byte, 1)
	sort.Slice(handles, func(i, j int) bool {
		return bytes.Compare(handles[i], handles[j]) < 0
	})
	l := len(handles)
	startKey := codec.EncodeBytes([]byte{}, handles[0])
	endKey := codec.EncodeBytes([]byte{}, handles[l-1])
	regions, err := paginateScanRegion(ctx, manager.splitCli, startKey, endKey, 128)
	if err != nil {
		return handles
	}
	startIdx := 0
	endIdx := 0
	batch := make([][]byte, len(handles))
	for _, region := range regions {
		handleKey := codec.EncodeBytes([]byte{}, handles[startIdx])
		if bytes.Compare(handleKey, region.Region.EndKey) >= 0 {
			continue
		}
		endIdx = startIdx
		for endIdx < l {
			handleKey := codec.EncodeBytes([]byte{}, handles[endIdx])
			if bytes.Compare(handleKey, region.Region.EndKey) < 0 {
				batch = append(batch, handles[endIdx])
			} else {
				break
			}
		}
		if err := manager.getValuesFromRegion(ctx, region, batch); err != nil {
			log.L().Error("failed to collect values from TiKV by handle, we will retry it again", zap.Error(err))
			retryHandles = append(retryHandles, batch...)
		}
		startIdx = endIdx
	}
	return retryHandles
}

func (manager *DuplicateManager) getValuesFromRegion(
	ctx context.Context,
	region *split.RegionInfo,
	handles [][]byte,
) error {
	kvclient, err := manager.getKvClient(ctx, region.Leader)
	if err != nil {
		return err
	}
	reqCtx := &kvrpcpb.Context{
		RegionId:    region.Region.GetId(),
		RegionEpoch: region.Region.GetRegionEpoch(),
		Peer:        region.Leader,
	}

	req := &kvrpc.BatchGetRequest{
		Context: reqCtx,
		Keys:    handles,
		Version: manager.ts,
	}
	resp, err := kvclient.KvBatchGet(ctx, req)
	if err != nil {
		return err
	}
	if resp.GetRegionError() != nil {
		return errors.Errorf("region error because of %v", resp.GetRegionError().GetMessage())
	}
	if resp.Error != nil {
		return errors.Errorf("key error")
	}
	for i := 0; i < maxRetryTimes; i++ {
		b := manager.db.NewBatch()
		opts := &pebble.WriteOptions{Sync: false}
		for _, kv := range resp.Pairs {
			b.Set(kv.Key, kv.Value, opts)
			if b.Count() > maxWriteBatchCount {
				err = b.Commit(opts)
				if err != nil {
					break
				} else {
					b.Reset()
				}
			}
		}
		if err == nil {
			err = b.Commit(opts)
		}
		if err == nil {
			return nil
		}
	}
	return err
}

func (manager *DuplicateManager) getDuplicateStream(ctx context.Context,
	region *split.RegionInfo,
	start []byte, end []byte) (sst.ImportSST_DuplicateDetectClient, error) {
	leader := region.Leader
	if leader == nil {
		leader = region.Region.GetPeers()[0]
	}

	cli, err := manager.getImportClient(ctx, leader)
	if err != nil {
		return nil, err
	}

	reqCtx := &kvrpcpb.Context{
		RegionId:    region.Region.GetId(),
		RegionEpoch: region.Region.GetRegionEpoch(),
		Peer:        leader,
	}
	req := &sst.DuplicateDetectRequest{
		Context:  reqCtx,
		StartKey: start,
		EndKey:   end,
		KeyOnly:  false,
	}
	stream, err := cli.DuplicateDetect(ctx, req)
	return stream, err
}

func (manager *DuplicateManager) getKvClient(ctx context.Context, peer *metapb.Peer) (tikv.TikvClient, error) {
	conn, err := manager.connPool.GetGrpcConn(ctx, peer.GetStoreId(), 1, func(ctx context.Context) (*grpc.ClientConn, error) {
		return manager.makeConn(ctx, peer.GetStoreId())
	})
	if err != nil {
		return nil, err
	}
	return tikv.NewTikvClient(conn), nil
}

func (manager *DuplicateManager) getImportClient(ctx context.Context, peer *metapb.Peer) (sst.ImportSSTClient, error) {
	conn, err := manager.connPool.GetGrpcConn(ctx, peer.GetStoreId(), 1, func(ctx context.Context) (*grpc.ClientConn, error) {
		return manager.makeConn(ctx, peer.GetStoreId())
	})
	if err != nil {
		return nil, err
	}
	return sst.NewImportSSTClient(conn), nil
}

func (manager *DuplicateManager) makeConn(ctx context.Context, storeID uint64) (*grpc.ClientConn, error) {
	store, err := manager.splitCli.GetStore(ctx, storeID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	opt := grpc.WithInsecure()
	if manager.tls.TLSConfig() != nil {
		opt = grpc.WithTransportCredentials(credentials.NewTLS(manager.tls.TLSConfig()))
	}
	ctx, cancel := context.WithTimeout(ctx, dialTimeout)

	bfConf := backoff.DefaultConfig
	bfConf.MaxDelay = gRPCBackOffMaxDelay
	// we should use peer address for tiflash. for tikv, peer address is empty
	addr := store.GetPeerAddress()
	if addr == "" {
		addr = store.GetAddress()
	}
	conn, err := grpc.DialContext(
		ctx,
		addr,
		opt,
		grpc.WithConnectParams(grpc.ConnectParams{Backoff: bfConf}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                gRPCKeepAliveTime,
			Timeout:             gRPCKeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	)
	cancel()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return conn, nil
}

func buildDuplicateRequests(tableInfo *model.TableInfo) ([]*DuplicateRequest, error) {
	reqs := make([]*DuplicateRequest, 0)
	req := buildTableRequest(tableInfo.ID)
	reqs = append(reqs, req...)
	for _, indexInfo := range tableInfo.Indices {
		if indexInfo.State != model.StatePublic {
			continue
		}
		req, err := buildIndexRequest(tableInfo.ID, indexInfo)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, req...)
	}
	return reqs, nil
}

func buildTableRequest(tableID int64) []*DuplicateRequest {
	ranges := ranger.FullIntRange(false)
	keysRanges := distsql.TableRangesToKVRanges(tableID, ranges, nil)
	reqs := make([]*DuplicateRequest, 1)
	for _, r := range keysRanges {
		r := &DuplicateRequest{
			start:   r.StartKey,
			end:     r.EndKey,
			indexID: 0,
			tableID: tableID,
		}
		reqs = append(reqs, r)
	}
	return reqs
}

func buildIndexRequest(tableID int64, indexInfo *model.IndexInfo) ([]*DuplicateRequest, error) {
	ranges := ranger.FullRange()
	keysRanges, err := distsql.IndexRangesToKVRanges(nil, tableID, indexInfo.ID, ranges, nil)
	if err != nil {
		return nil, err
	}
	reqs := make([]*DuplicateRequest, 1)
	for _, r := range keysRanges {
		r := &DuplicateRequest{
			start:     r.StartKey,
			end:       r.EndKey,
			indexID:   indexInfo.ID,
			tableID:   tableID,
			indexInfo: indexInfo,
		}
		reqs = append(reqs, r)
	}
	return reqs, nil
}
