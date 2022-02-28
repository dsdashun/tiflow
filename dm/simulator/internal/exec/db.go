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
	"database/sql"
	"sync"
	"sync/atomic"
)

type dbExec struct {
	db       *sql.DB
	txs      sync.Map
	lastTxID uint32
}

func NewDBExec(db *sql.DB) *dbExec {
	exec := &dbExec{
		db: db,
	}
	return exec
}

func (e *dbExec) BeginTx(ctx context.Context) (uint32, error) {
	var err error
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	var newTxID uint32 = atomic.AddUint32(&e.lastTxID, 1)
	if newTxID == 0 {
		newTxID = atomic.AddUint32(&e.lastTxID, 1)
	}
	for {
		if _, ok := e.txs.Load(newTxID); !ok {
			break
		}
		newTxID = atomic.AddUint32(&e.lastTxID, 1)
	}
	e.txs.Store(newTxID, tx)
	return newTxID, nil
}

func (e *dbExec) Commit(txID uint32) error {
	txVal, ok := e.txs.Load(txID)
	if !ok {
		return ErrTxIDNotFound
	}
	if err := txVal.(*sql.Tx).Commit(); err != nil {
		return err
	}
	e.txs.Delete(txID)
	return nil
}

func (e *dbExec) Rollback(txID uint32) error {
	txVal, ok := e.txs.Load(txID)
	if !ok {
		return ErrTxIDNotFound
	}
	if err := txVal.(*sql.Tx).Rollback(); err != nil {
		return err
	}
	e.txs.Delete(txID)
	return nil
}

func (e *dbExec) ExecContext(ctx context.Context, txID uint32, query string, args ...interface{}) (sql.Result, error) {
	txVal, ok := e.txs.Load(txID)
	if !ok {
		return nil, ErrTxIDNotFound
	}
	tx := txVal.(*sql.Tx)
	return tx.ExecContext(ctx, query, args...)
}
