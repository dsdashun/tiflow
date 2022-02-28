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

package exec

import (
	"context"
	"sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestParallelTrx(t *testing.T) {
	var err error
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	theDBExec := NewDBExec(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mock.MatchExpectationsInOrder(false)
	var mockLock sync.Mutex
	beginTxFn := func() uint32 {
		mockLock.Lock()
		defer mockLock.Unlock()
		mock.ExpectBegin()
		txID, err := theDBExec.BeginTx(ctx)
		assert.Nil(t, err)
		return txID
	}
	commitFn := func(txID uint32) {
		mockLock.Lock()
		defer mockLock.Unlock()
		mock.ExpectCommit()
		err := theDBExec.Commit(txID)
		assert.Nil(t, err)
	}
	rollbackFn := func(txID uint32) {
		mockLock.Lock()
		defer mockLock.Unlock()
		mock.ExpectRollback()
		err := theDBExec.Rollback(txID)
		assert.Nil(t, err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			txID1 := beginTxFn()
			txID2 := beginTxFn()
			t.Logf("tx ID1: %d; tx ID2: %d\n", txID1, txID2)
			commitFn(txID1)
			t.Logf("tx ID %d committed\n", txID1)
			rollbackFn(txID2)
			t.Logf("tx ID %d rollbacked\n", txID2)
		}()
	}
	wg.Wait()
}
