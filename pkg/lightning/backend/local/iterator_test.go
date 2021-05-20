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
	"math/rand"
	"path/filepath"
	"sort"
	"time"

	"github.com/cockroachdb/pebble"
	. "github.com/pingcap/check"

	"github.com/pingcap/br/pkg/lightning/common"
)

type iteratorSuite struct{}

var _ = Suite(&iteratorSuite{})

func (s *iteratorSuite) TestIterator(c *C) {
	var pairs []common.KvPair
	// Unique pairs.
	for i := 0; i < 20; i++ {
		pairs = append(pairs, common.KvPair{
			Key:    randBytes(32),
			Val:    randBytes(128),
			Offset: int64(i * 1234),
		})
	}
	// Duplicate pairs which repeat the same key twice.
	for i := 20; i < 40; i++ {
		key := randBytes(32)
		pairs = append(pairs, common.KvPair{
			Key:    key,
			Val:    randBytes(128),
			Offset: int64(i * 1234),
		})
		pairs = append(pairs, common.KvPair{
			Key:    key,
			Val:    randBytes(128),
			Offset: int64(i * 1235),
		})
	}
	// Duplicate pairs which repeat the same key three times.
	for i := 40; i < 50; i++ {
		key := randBytes(32)
		pairs = append(pairs, common.KvPair{
			Key:    key,
			Val:    randBytes(128),
			Offset: int64(i * 1234),
		})
		pairs = append(pairs, common.KvPair{
			Key:    key,
			Val:    randBytes(128),
			Offset: int64(i * 1235),
		})
		pairs = append(pairs, common.KvPair{
			Key:    key,
			Val:    randBytes(128),
			Offset: int64(i * 1236),
		})
	}

	// Find duplicates from the generated pairs.
	var duplicatePairs []common.KvPair
	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare(pairs[i].Key, pairs[j].Key) < 0
	})
	uniqueKeys := make([][]byte, 0)
	for i := 0; i < len(pairs); {
		j := i + 1
		for j < len(pairs) && bytes.Equal(pairs[j-1].Key, pairs[j].Key) {
			j++
		}
		uniqueKeys = append(uniqueKeys, pairs[i].Key)
		if i+1 == j {
			i++
			continue
		}
		for k := i; k < j; k++ {
			duplicatePairs = append(duplicatePairs, pairs[k])
		}
		i = j
	}

	// Write pairs to db after shuffling the pairs.
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})
	storeDir := c.MkDir()
	db, err := pebble.Open(filepath.Join(storeDir, "kv"), &pebble.Options{})
	c.Assert(err, IsNil)
	wb := db.NewBatch()
	for _, p := range pairs {
		key := EncodeKeySuffix(nil, p.Key, []byte("table.sql"), p.Offset)
		c.Assert(wb.Set(key, p.Val, nil), IsNil)
	}
	c.Assert(wb.Commit(pebble.Sync), IsNil)

	duplicateDBPath := filepath.Join(storeDir, "duplicate-kv")
	engineFile := &File{
		ctx:             context.Background(),
		db:              db,
		duplicateDBPath: duplicateDBPath,
	}
	iter := newDuplicateIterator(engineFile, &pebble.IterOptions{})
	sort.Slice(pairs, func(i, j int) bool {
		key1 := EncodeKeySuffix(nil, pairs[i].Key, []byte("table.sql"), pairs[i].Offset)
		key2 := EncodeKeySuffix(nil, pairs[j].Key, []byte("table.sql"), pairs[j].Offset)
		return bytes.Compare(key1, key2) < 0
	})

	// Verify first pair.
	c.Assert(iter.First(), IsTrue)
	c.Assert(iter.Valid(), IsTrue)
	c.Assert(iter.Key(), BytesEquals, pairs[0].Key)
	c.Assert(iter.Value(), BytesEquals, pairs[0].Val)

	// Verify last pair.
	c.Assert(iter.Last(), IsTrue)
	c.Assert(iter.Valid(), IsTrue)
	c.Assert(iter.Key(), BytesEquals, pairs[len(pairs)-1].Key)
	c.Assert(iter.Value(), BytesEquals, pairs[len(pairs)-1].Val)

	// Iterate all keys and check the count of unique keys.
	for iter.First(); iter.Valid(); iter.Next() {
		c.Assert(iter.Key(), BytesEquals, uniqueKeys[0])
		uniqueKeys = uniqueKeys[1:]
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(len(uniqueKeys), Equals, 0)
	c.Assert(iter.Close(), IsNil)
	c.Assert(engineFile.Close(), IsNil)

	// Check duplicates detected by duplicate iterator.
	duplicateDB, err := pebble.Open(duplicateDBPath, &pebble.Options{})
	c.Assert(err, IsNil)
	iter = duplicateDB.NewIter(&pebble.IterOptions{})
	var detectedPairs []common.KvPair
	for iter.First(); iter.Valid(); iter.Next() {
		key, err := DecodeKeySuffix(nil, iter.Key())
		c.Assert(err, IsNil)
		detectedPairs = append(detectedPairs, common.KvPair{
			Key: key,
			Val: append([]byte{}, iter.Value()...),
		})
	}
	c.Assert(iter.Error(), IsNil)
	c.Assert(iter.Close(), IsNil)
	c.Assert(duplicateDB.Close(), IsNil)
	c.Assert(len(detectedPairs), Equals, len(duplicatePairs))

	sort.Slice(duplicatePairs, func(i, j int) bool {
		keyCmp := bytes.Compare(duplicatePairs[i].Key, duplicatePairs[j].Key)
		return keyCmp < 0 || keyCmp == 0 && bytes.Compare(duplicatePairs[i].Val, duplicatePairs[j].Val) < 0
	})
	sort.Slice(detectedPairs, func(i, j int) bool {
		keyCmp := bytes.Compare(detectedPairs[i].Key, detectedPairs[j].Key)
		return keyCmp < 0 || keyCmp == 0 && bytes.Compare(detectedPairs[i].Val, detectedPairs[j].Val) < 0
	})
	for i := 0; i < len(detectedPairs); i++ {
		c.Assert(detectedPairs[i].Key, BytesEquals, duplicatePairs[i].Key)
		c.Assert(detectedPairs[i].Val, BytesEquals, duplicatePairs[i].Val)
	}
}
