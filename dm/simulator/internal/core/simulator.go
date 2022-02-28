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
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/pingcap/tiflow/dm/simulator/internal/exec"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type simulatorImpl struct {
	sqlExec          exec.SQLExecutor
	dmlGen           sqlgen.DMLSQLGenerator
	mcp              *sqlgen.ModificationCandidatePool
	totalExecutedTrx uint64
}

func NewSimulatorImpl(sqlExec exec.SQLExecutor) *simulatorImpl {
	return &simulatorImpl{
		dmlGen:  sqlgen.NewDMLSQLGenerator(),
		mcp:     sqlgen.NewModificationCandidatePool(),
		sqlExec: sqlExec,
	}
}

func (s *simulatorImpl) DoSimulation(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// continue
		}
		s.simulateTrx(ctx)
	}
}

func (s *simulatorImpl) prepareData(ctx context.Context) error {
	var err error
	txID, err := s.sqlExec.BeginTx(ctx)
	if err != nil {
		return err
	}
	_, err = s.sqlExec.ExecContext(ctx, txID, "TRUNCATE TABLE games.members")
	if err != nil {
		return err
	}
	for i := 0; i < 4096; i++ {
		sql, uk, err := s.dmlGen.GenInsertRow()
		_, err = s.sqlExec.ExecContext(ctx, txID, sql)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = s.mcp.AddUK(uk)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return s.sqlExec.Commit(txID)
}

func (s *simulatorImpl) simulateTrx(ctx context.Context) error {
	var err error
	txID, err := s.sqlExec.BeginTx(ctx)
	if err != nil {
		return err
	}
	dmlType := randType()
	switch dmlType {
	case sqlgen.INSERT_DMLType:
		_, err = s.simulateInsert(ctx, txID)
		if err != nil {
			fmt.Println(err)
			s.sqlExec.Rollback(txID)
			return err
		}
	case sqlgen.DELETE_DMLType:
		err = s.simulateDelete(ctx, txID)
		if err != nil {
			fmt.Println(err)
			s.sqlExec.Rollback(txID)
			return err
		}
	default:
		err = s.simulateUpdate(ctx, txID)
		if err != nil {
			fmt.Println(err)
			s.sqlExec.Rollback(txID)
			return err
		}
	}
	err = s.sqlExec.Commit(txID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	atomic.AddUint64(&s.totalExecutedTrx, 1)
	return nil
}

func (s *simulatorImpl) simulateInsert(ctx context.Context, txID uint32) (*sqlgen.UniqueKey, error) {
	var err error
	sql, uk, err := s.dmlGen.GenInsertRow()
	_, err = s.sqlExec.ExecContext(ctx, txID, sql)
	if err != nil {
		return nil, err
	}
	err = s.mcp.AddUK(uk)
	if err != nil {
		return nil, err
	}
	return uk, nil
}

func (s *simulatorImpl) simulateDelete(ctx context.Context, txID uint32) error {
	var err error
	uk := s.mcp.NextUK()
	if uk == nil {
		return ErrNoMCPData
	}
	sql, err := s.dmlGen.GenDeleteRow(uk)
	if err != nil {
		return err
	}
	_, err = s.sqlExec.ExecContext(ctx, txID, sql)
	if err != nil {
		return err
	}
	err = s.mcp.DeleteUK(uk)
	if err != nil {
		return err
	}
	return nil
}

func (s *simulatorImpl) simulateUpdate(ctx context.Context, txID uint32) error {
	var err error
	uk := s.mcp.NextUK()
	if uk == nil {
		return ErrNoMCPData
	}
	sql, err := s.dmlGen.GenUpdateRow(uk)
	if err != nil {
		return err
	}
	_, err = s.sqlExec.ExecContext(ctx, txID, sql)
	if err != nil {
		return err
	}
	return nil
}

func randType() sqlgen.DMLType {
	randNum := rand.Int() % 4
	switch randNum {
	case 0:
		return sqlgen.INSERT_DMLType
	case 1:
		return sqlgen.DELETE_DMLType
	default:
		return sqlgen.UPDATE_DMLType
	}
}
