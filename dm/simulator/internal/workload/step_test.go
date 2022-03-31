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

package workload

import (
	"context"
	"math/rand"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type testWorkloadStepSuite struct {
	suite.Suite
	tableConfig *config.TableConfig
	theMCP      *mcp.ModificationCandidatePool
}

func (s *testWorkloadStepSuite) SetupSuite() {
	s.Require().Nil(log.InitLogger(&log.Config{}))
	s.tableConfig = &config.TableConfig{
		DatabaseName: "games",
		TableName:    "members",
		Columns: []*config.ColumnDefinition{
			{
				ColumnName: "id",
				DataType:   "int",
			},
			{
				ColumnName: "name",
				DataType:   "varchar",
			},
			{
				ColumnName: "age",
				DataType:   "int",
			},
			{
				ColumnName: "team_id",
				DataType:   "int",
			},
		},
		UniqueKeyColumnNames: []string{"id"},
	}
	s.theMCP = mcp.NewModificationCandidatePool(8192)
	for i := 0; i < 100; i++ {
		s.Require().Nil(
			s.theMCP.AddUK(mcp.NewUniqueKey(-1, map[string]interface{}{
				"id": rand.Int(),
			})),
		)
	}
}

func (s *testWorkloadStepSuite) TestBasic() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(s.tableConfig)
	theInsertStep := &InsertStep{
		sqlGen: sqlGen,
	}
	theUpdateStep := &UpdateStep{
		sqlGen: sqlGen,
	}
	theDeleteStep := &DeleteStep{
		sqlGen: sqlGen,
	}
	theRandomStep := &RandomDMLStep{
		sqlGen: sqlGen,
	}

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	s.Require().Nil(err)

	sctx := &DMLWorkloadStepContext{
		tx:       tx,
		ctx:      ctx,
		mcp:      s.theMCP,
		rowRefs:  make(map[string]*mcp.UniqueKey),
		addedUKs: make(map[string]map[*mcp.UniqueKey]struct{}),
	}
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theInsertStep.Execute(sctx))

	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theUpdateStep.Execute(sctx))

	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theDeleteStep.Execute(sctx))

	mock.ExpectExec("^(INSERT|UPDATE|DELETE) (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theRandomStep.Execute(sctx))

	mock.ExpectCommit()
	s.Require().Nil(tx.Commit())
}

func (s *testWorkloadStepSuite) TestAssignmentReference() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(s.tableConfig)

	theInsertStep := &InsertStep{
		sqlGen: sqlGen,
	}

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	s.Require().Nil(err)

	sctx := &DMLWorkloadStepContext{
		tx:       tx,
		ctx:      ctx,
		mcp:      s.theMCP,
		rowRefs:  make(map[string]*mcp.UniqueKey),
		addedUKs: make(map[string]map[*mcp.UniqueKey]struct{}),
	}
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theInsertStep.Execute(sctx))
	s.Require().Equal(0, len(sctx.rowRefs), "there should be no assigned rows")

	assignedRowID := "@abc01"
	theInsertStep.assignedRowID = assignedRowID

	mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theInsertStep.Execute(sctx))
	assignedUK, ok := sctx.rowRefs[assignedRowID]
	s.Require().Truef(ok, "%s should be assigned", assignedRowID)
	s.T().Logf("%s assigned with the UK: %v\n", assignedRowID, assignedUK)

	// normal update
	theUpdateStep := &UpdateStep{
		sqlGen: sqlGen,
	}
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theUpdateStep.Execute(sctx))

	// update use the assigned UK
	theUpdateStep.inputRowID = assignedRowID
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theUpdateStep.Execute(sctx))

	// update use an non-existing UK
	theUpdateStep.inputRowID = "@NOT_EXISTING"
	s.Require().NotNil(theUpdateStep.Execute(sctx), "update a non-existing row should have error")
	s.T().Logf("updating a non-existing rowID get the following error: %v\n", err)

	// double reference the same UK
	assignedRowID2 := "@abc02"
	theUpdateStep.inputRowID = assignedRowID
	theUpdateStep.assignmentRowID = assignedRowID2
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theUpdateStep.Execute(sctx))
	assignedUK2, ok := sctx.rowRefs[assignedRowID2]
	s.Require().Truef(ok, "%s should be assigned", assignedRowID2)
	s.T().Logf("%s assigned with the UK: %v\n", assignedRowID2, assignedUK2)
	s.Require().Equal(assignedUK, assignedUK2, "the two assignment should be the same")

	// delete double-refferred row
	theDeleteStep := &DeleteStep{
		sqlGen:     sqlGen,
		inputRowID: assignedRowID2,
	}
	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theDeleteStep.Execute(sctx))
	_, ok = sctx.rowRefs[assignedRowID2]
	s.Require().Falsef(ok, "%s should not be assigned", assignedRowID2)
	s.T().Logf("after delete %s: %v\n", assignedRowID2, assignedUK)

	// assign another row to the row ID
	theUpdateStep.assignmentRowID = assignedRowID
	theUpdateStep.inputRowID = ""
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theUpdateStep.Execute(sctx))
	anotherAssignedUK, ok := sctx.rowRefs[assignedRowID]
	s.Require().Truef(ok, "%s should be assigned", assignedRowID)
	s.T().Logf("%s assigned with the UK: %v\n", assignedRowID, anotherAssignedUK)

	// delete non-existing row-ref
	theDeleteStep.inputRowID = "@NOT_EXISTING"
	s.Require().NotNil(theDeleteStep.Execute(sctx), "delete a non-existing row should have error")
	s.T().Logf("deleting a non-existing rowID get the following error: %v\n", err)

	// delete existing row-ref
	theDeleteStep.inputRowID = assignedRowID
	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theDeleteStep.Execute(sctx))
	_, ok = sctx.rowRefs[assignedRowID]
	s.Require().Falsef(ok, "%s should be unassigned", assignedRowID)

	// normal deletion
	theDeleteStep.inputRowID = ""
	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	s.Require().Nil(theDeleteStep.Execute(sctx))

	mock.ExpectCommit()
	s.Require().Nil(tx.Commit())
}

func TestWorkloadStepSuite(t *testing.T) {
	suite.Run(t, &testWorkloadStepSuite{})
}
