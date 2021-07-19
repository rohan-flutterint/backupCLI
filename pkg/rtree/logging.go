// Copyright 2021 PingCAP, Inc. Licensed under Apache-2.0.

package rtree

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/docker/go-units"
	"github.com/pingcap/br/pkg/logutil"
	"github.com/pingcap/br/pkg/redact"
)

// String formats a range to a string.
func (rg Range) String() string {
	return fmt.Sprintf("[%s, %s)", redact.Key(rg.StartKey), redact.Key(rg.EndKey))
}

// ZapRanges make zap fields for logging Range slice.
func ZapRanges(ranges []Range) zapcore.Field {
	return zap.Object("ranges", rangesMarshaler(ranges))
}

type rangesMarshaler []Range

func (rs rangesMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, r := range rs {
		encoder.AppendString(r.String())
	}
	return nil
}

func (rs rangesMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	total := len(rs)
	encoder.AddInt("total", total)
	elements := make([]string, 0, total)
	for _, r := range rs {
		elements = append(elements, r.String())
	}
	_ = encoder.AddArray("ranges", logutil.AbbreviatedArrayMarshaler(elements))

	totalKV := uint64(0)
	totalBytes := uint64(0)
	totalSize := uint64(0)
	totalFile := 0
	for _, r := range rs {
		for _, f := range r.Files {
			totalKV += f.GetTotalKvs()
			totalBytes += f.GetTotalBytes()
			totalSize += f.GetSize_()
		}
		totalFile += len(r.Files)
	}

	encoder.AddInt("file-count", totalFile)
	encoder.AddUint64("kv-paris-count", totalKV)
	encoder.AddString("after-compress-size", units.HumanSize(float64(totalBytes)))
	encoder.AddString("data-size", units.HumanSize(float64(totalSize)))
	return nil
}
