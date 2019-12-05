package backup

import (
	"context"
	"testing"
	"time"

	"github.com/pingcap/br/pkg/conn"
	"github.com/pingcap/br/pkg/utils"

	. "github.com/pingcap/check"
	"github.com/pingcap/parser/model"
	"github.com/pingcap/tidb/store/mockstore/mocktikv"
)

type testBackup struct {
	ctx    context.Context
	cancel context.CancelFunc

	backupClient *Client
}

var _ = Suite(&testBackup{})

func TestT(t *testing.T) {
	TestingT(t)
}

func (r *testBackup) SetUpSuite(c *C) {
	mockPDClient := mocktikv.NewPDClient(mocktikv.NewCluster())
	r.ctx, r.cancel = context.WithCancel(context.Background())
	mockMgr := &conn.Mgr{
		PDClient: mockPDClient,
	}
	mockMgr.SetPDHTTP([]string{"test"}, nil)
	r.backupClient = &Client{
		clusterID: mockPDClient.GetClusterID(r.ctx),
		mgr:       mockMgr,
		pdClient:  mockPDClient,
	}
}

func (r *testBackup) TestGetTS(c *C) {
	var (
		err error
		// mockPDClient' physical ts and current ts will have deviation
		// so make this deviation tolerance 100ms
		deviation = 100
	)

	// timeago not valid
	timeAgo := "invalid"
	_, err = r.backupClient.GetTS(r.ctx, timeAgo)
	c.Assert(err, ErrorMatches, "time: invalid duration invalid")

	// timeago not work
	timeAgo = ""
	expectedDuration := 0
	currentTs := time.Now().UnixNano() / int64(time.Millisecond)
	ts, err := r.backupClient.GetTS(r.ctx, timeAgo)
	c.Assert(err, IsNil)
	pdTs := utils.DecodeTs(ts).Physical
	duration := int(currentTs - pdTs)
	c.Assert(duration, Greater, expectedDuration-deviation)
	c.Assert(duration, Less, expectedDuration+deviation)

	// timeago = "1.5m"
	timeAgo = "1.5m"
	expectedDuration = 90000
	currentTs = time.Now().UnixNano() / int64(time.Millisecond)
	ts, err = r.backupClient.GetTS(r.ctx, timeAgo)
	c.Assert(err, IsNil)
	pdTs = utils.DecodeTs(ts).Physical
	duration = int(currentTs - pdTs)
	c.Assert(duration, Greater, expectedDuration-deviation)
	c.Assert(duration, Less, expectedDuration+deviation)

	// timeago = "-1m"
	timeAgo = "-1m"
	expectedDuration = -60000
	currentTs = time.Now().UnixNano() / int64(time.Millisecond)
	ts, err = r.backupClient.GetTS(r.ctx, timeAgo)
	c.Assert(err, IsNil)
	pdTs = utils.DecodeTs(ts).Physical
	duration = int(currentTs - pdTs)
	c.Assert(duration, Greater, expectedDuration-deviation)
	c.Assert(duration, Less, expectedDuration+deviation)

	// timeago = "1000000h" exceed GCSafePoint
	// because GCSafePoint in mockPDClient is 0
	timeAgo = "1000000h"
	_, err = r.backupClient.GetTS(r.ctx, timeAgo)
	c.Assert(err, ErrorMatches, "given backup time exceed GCSafePoint")
}

func (r *testBackup) TestBuildTableRange(c *C) {
	type Case struct {
		ids []int64
		trs []tableRange
	}
	cases := []Case{
		{ids: []int64{1}, trs: []tableRange{{startID: 1, endID: 2}}},
		{ids: []int64{1, 2, 3}, trs: []tableRange{
			{startID: 1, endID: 2},
			{startID: 2, endID: 3},
			{startID: 3, endID: 4},
		}},
		{ids: []int64{1, 3}, trs: []tableRange{
			{startID: 1, endID: 2},
			{startID: 3, endID: 4},
		}},
	}
	for _, cs := range cases {
		c.Log(cs)
		tbl := &model.TableInfo{Partition: &model.PartitionInfo{Enable: true}}
		for _, id := range cs.ids {
			tbl.Partition.Definitions = append(tbl.Partition.Definitions,
				model.PartitionDefinition{ID: id})
		}
		ranges := buildTableRanges(tbl)
		c.Assert(ranges, DeepEquals, cs.trs)
	}

	tbl := &model.TableInfo{ID: 7}
	ranges := buildTableRanges(tbl)
	c.Assert(ranges, DeepEquals, []tableRange{{startID: 7, endID: 8}})

}
