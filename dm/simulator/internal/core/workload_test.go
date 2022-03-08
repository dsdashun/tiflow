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
	"math/rand"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type testWorkloadSimulatorSuite struct {
	suite.Suite
	tableConfig *config.TableConfig
	mcpMap      map[string]*sqlgen.ModificationCandidatePool
}

func (s *testWorkloadSimulatorSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
	s.tableConfig = &config.TableConfig{
		TableID:      "members",
		DatabaseName: "games",
		TableName:    "members",
		Columns: []*config.ColumnDefinition{
			&config.ColumnDefinition{
				ColumnName: "id",
				DataType:   "int",
				DataLen:    11,
			},
			&config.ColumnDefinition{
				ColumnName: "name",
				DataType:   "varchar",
				DataLen:    255,
			},
			&config.ColumnDefinition{
				ColumnName: "age",
				DataType:   "int",
				DataLen:    11,
			},
			&config.ColumnDefinition{
				ColumnName: "team_id",
				DataType:   "int",
				DataLen:    11,
			},
		},
		UniqueKeyColumnNames: []string{"id"},
	}
	s.mcpMap = make(map[string]*sqlgen.ModificationCandidatePool)
}

func (s *testWorkloadSimulatorSuite) SetupTest() {
	mcp := sqlgen.NewModificationCandidatePool()
	recordCount := 128
	for i := 0; i < recordCount; i++ {
		mcp.AddUK(&sqlgen.UniqueKey{
			RowID: -1,
			Value: map[string]interface{}{
				"id": rand.Int(),
			},
		})
	}
	s.mcpMap["members"] = mcp
}

func mockSingleDMLTrx(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectExec("^(INSERT|UPDATE|DELETE) (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func (s *testWorkloadSimulatorSuite) TestBasic() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	theSimulator, err := NewWorkloadSimulatorImpl(
		map[string]*config.TableConfig{
			"members": s.tableConfig,
		},
		"RANDOM-DML members;",
	)
	assert.Nil(s.T(), err)
	for i := 0; i < 100; i++ {
		mockSingleDMLTrx(mock)
		err = theSimulator.SimulateTrx(ctx, db, s.mcpMap)
		assert.Nil(s.T(), err)
	}
	s.T().Logf("total executed trx: %d\n", theSimulator.totalExecutedTrx)
}

func (s *testWorkloadSimulatorSuite) TestParallelSimulation() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerCount := 4
	workerFn := func() error {
		var err error
		db, mock, err := sqlmock.New()
		if err != nil {
			s.T().Logf("open testing DB failed: %v\n", err)
			return err
		}
		theSimulator, err := NewWorkloadSimulatorImpl(
			map[string]*config.TableConfig{
				"members": s.tableConfig,
			},
			"RANDOM-DML members;",
		)
		if err != nil {
			return err
		}

		for i := 0; i < 100; i++ {
			mockSingleDMLTrx(mock)
			err = theSimulator.SimulateTrx(ctx, db, s.mcpMap)
			if err != nil {
				return err
			}
		}
		return nil
	}
	resultCh := make(chan error, workerCount)
	defer close(resultCh)
	for i := 0; i < workerCount; i++ {
		go func() {
			err := workerFn()
			resultCh <- err
		}()
	}
	for i := 0; i < workerCount; i++ {
		err := <-resultCh
		assert.Nil(s.T(), err)
	}
}

func TestWorkloadSimulatorSuite(t *testing.T) {
	suite.Run(t, &testWorkloadSimulatorSuite{})
}
