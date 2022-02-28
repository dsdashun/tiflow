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

package core

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/pingcap/tiflow/dm/simulator/internal/exec"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type dummyExec struct {
	totalExecuted    int
	trxTotalExecuted int
}

func (e *dummyExec) BeginTx(ctx context.Context) (uint32, error) {
	return 1, nil
}

func (e *dummyExec) Commit(txID uint32) error {
	e.trxTotalExecuted++
	return nil
}

func (e *dummyExec) Rollback(txID uint32) error {
	return nil
}

func (e *dummyExec) ExecContext(ctx context.Context, txID uint32, query string, args ...interface{}) (sql.Result, error) {
	// fmt.Printf("executing: %s; arguments: %v\n", query, args)
	e.totalExecuted++
	return nil, nil
}

func TestSimulatorBasic(t *testing.T) {
	// theExec := &dummyExec{}
	db, err := sql.Open("mysql", "root:guanliyuanmima@tcp(127.0.0.1:13306)/games")
	// db, err := sql.Open("mysql", "root:@tcp(rms-staging.pingcap.net:31469)/games")
	if err != nil {
		t.Fatalf("open testing DB failed: %v\n", err)
	}
	theExec := exec.NewDBExec(db)
	s := NewSimulatorImpl(theExec)
	err = s.prepareData(context.Background())
	assert.Nil(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.DoSimulation(ctx)
	t.Logf("total executed trx: %d\n", s.totalExecutedTrx)
}

func TestSingleSimulation(t *testing.T) {
	var (
		err error
		uk  *sqlgen.UniqueKey
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	theExec := &dummyExec{}
	/*
		db, err := sql.Open("mysql", "root:guanliyuanmima@tcp(127.0.0.1:13306)/games")
		if err != nil {
			t.Fatalf("open testing DB failed: %v\n", err)
		}
		theExec := exec.NewDBExec(db)
	*/
	s := NewSimulatorImpl(theExec)
	// err = s.prepareData(ctx)
	assert.Nil(t, err)
	txID, err := s.sqlExec.BeginTx(ctx)
	assert.Nil(t, err)
	uk, err = s.simulateInsert(ctx, txID)
	assert.Nil(t, err)
	t.Logf("new UK: %v\n", uk)
	err = s.simulateUpdate(ctx, txID)
	assert.Nil(t, err)
	err = s.simulateDelete(ctx, txID)
	assert.Nil(t, err)
	err = s.sqlExec.Commit(txID)
	assert.Nil(t, err)
}
