// Copyright 2022 PingCAP, Inc.
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

package sqlgen

import (
	"math/rand"
	"sync"
	"time"
)

var (
	initRandSeedOnce sync.Once
	gMCPRand         *rand.Rand
)

func init() {
	initRandSeedOnce.Do(func() {
		gMCPRand = rand.New(rand.NewSource(time.Now().Unix()))
	})
}

type ModificationCandidatePool struct {
	sync.RWMutex
	keyPool []*UniqueKey
}

func NewModificationCandidatePool() *ModificationCandidatePool {
	theKeyPool := make([]*UniqueKey, 0, 8192)
	return &ModificationCandidatePool{
		keyPool: theKeyPool,
	}
}

func (mcp *ModificationCandidatePool) PreparePool() {
	for i := 0; i < 4096; i++ {
		currentLen := len(mcp.keyPool)
		mcp.keyPool = append(mcp.keyPool, &UniqueKey{
			RowID: currentLen,
			Value: map[string]interface{}{
				"id": i,
			},
		})
	}
}

func (mcp *ModificationCandidatePool) NextUK() *UniqueKey {
	mcp.RLock()
	defer mcp.RUnlock()
	if len(mcp.keyPool) == 0 {
		return nil
	}
	idx := gMCPRand.Intn(len(mcp.keyPool))
	result := mcp.keyPool[idx].Clone()
	result.RowID = idx
	return result
}

func (mcp *ModificationCandidatePool) GetUKByRowID(rowID int) *UniqueKey {
	mcp.RLock()
	defer mcp.RUnlock()
	if rowID >= 0 && rowID < len(mcp.keyPool) {
		return mcp.keyPool[rowID]
	} else {
		return nil
	}
}

func (mcp *ModificationCandidatePool) Len() int {
	mcp.RLock()
	defer mcp.RUnlock()
	return len(mcp.keyPool)
}

func (mcp *ModificationCandidatePool) AddUK(uk *UniqueKey) error {
	mcp.Lock()
	defer mcp.Unlock()
	if len(mcp.keyPool)+1 > cap(mcp.keyPool) {
		return ErrMCPCapacityFull
	}
	currentLen := len(mcp.keyPool)
	clonedUK := uk.Clone()
	clonedUK.RowID = currentLen
	mcp.keyPool = append(mcp.keyPool, clonedUK)
	return nil
}

func (mcp *ModificationCandidatePool) DeleteUK(uk *UniqueKey) error {
	var (
		deletedUK *UniqueKey
		deleteIdx int = -1
	)
	if uk == nil {
		return nil
	}
	mcp.Lock()
	defer mcp.Unlock()
	if uk.RowID >= 0 {
		if uk.RowID >= len(mcp.keyPool) {
			return ErrInvalidRowID
		}
		deletedUK = mcp.keyPool[uk.RowID]
		deleteIdx = uk.RowID
	} else {
		for i, searchUK := range mcp.keyPool {
			if searchUK.IsValueEqual(uk) {
				deletedUK = searchUK
				deleteIdx = i
				break
			}
		}
	}
	if deleteIdx < 0 {
		return ErrDeleteUKNotFound
	}
	curLen := len(mcp.keyPool)
	lastUK := mcp.keyPool[curLen-1].Clone()
	lastUK.RowID = deleteIdx
	mcp.keyPool[deleteIdx] = lastUK
	mcp.keyPool[curLen-1] = deletedUK
	mcp.keyPool = mcp.keyPool[:curLen-1]
	return nil
}

func (mcp *ModificationCandidatePool) Reset() {
	mcp.keyPool = mcp.keyPool[:0]
}
