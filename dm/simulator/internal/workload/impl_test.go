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
	"fmt"
	"math/rand"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
)

type testWorkloadSimulatorSuite struct {
	suite.Suite
	mcpMap map[string]*mcp.ModificationCandidatePool
}

func (s *testWorkloadSimulatorSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
	s.mcpMap = make(map[string]*mcp.ModificationCandidatePool)
}

func newTestMCP(recordCount int, tblConfig *config.TableConfig) (*mcp.ModificationCandidatePool, error) {
	theMCP := mcp.NewModificationCandidatePool(8192)
	colDefMap := make(map[string]*config.ColumnDefinition)
	for _, colDef := range tblConfig.Columns {
		colDefMap[colDef.ColumnName] = colDef
	}
	for i := 0; i < recordCount; i++ {
		ukValue := make(map[string]interface{})
		for _, ukColName := range tblConfig.UniqueKeyColumnNames {
			colDef, ok := colDefMap[ukColName]
			if !ok {
				errMsg := "cannot find the column definition"
				log.L().Error(errMsg, zap.String("col_name", ukColName))
				return nil, errors.New(errMsg)
			}
			switch colDef.DataType {
			case "int":
				ukValue[ukColName] = rand.Int()
			case "varchar":
				ukValue[ukColName] = fmt.Sprintf("val-%d", rand.Int())
			default:
				errMsg := "unsupported data type"
				log.L().Error(errMsg, zap.String("data_type", colDef.DataType))
				return nil, errors.New(errMsg)
			}
		}
		if err := theMCP.AddUK(mcp.NewUniqueKey(-1, ukValue)); err != nil {
			return nil, errors.Annotate(err, "add UK to the MCP error")
		}
	}
	return theMCP, nil
}

func (s *testWorkloadSimulatorSuite) SetupTest() {
	theMCP := mcp.NewModificationCandidatePool(8192)
	recordCount := 128
	for i := 0; i < recordCount; i++ {
		assert.Nil(s.T(),
			theMCP.AddUK(mcp.NewUniqueKey(-1, map[string]interface{}{
				"id": rand.Int(),
			})),
		)
	}
	s.mcpMap["members"] = theMCP
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
			"members": config.NewTemplateTableConfigForTest(),
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

func mockInsertStep(mock sqlmock.Sqlmock, tblConfig *config.TableConfig) {
	var colNames []string
	for _, colDef := range tblConfig.Columns {
		colNames = append(colNames, fmt.Sprintf("`%s`", colDef.ColumnName))
	}

	mock.ExpectExec(
		fmt.Sprintf("^INSERT INTO `%s`\\.`%s` \\(((%s),){%d}(%s)\\) VALUES .+",
			tblConfig.DatabaseName,
			tblConfig.TableName,
			strings.Join(colNames, "|"),
			len(colNames)-1,
			strings.Join(colNames, "|"),
		),
	).WillReturnResult(sqlmock.NewResult(0, 1))
}

func mockUpdateStep(mock sqlmock.Sqlmock, tblConfig *config.TableConfig) {
	var (
		nonUKColNames    []string
		quotedUKColNames []string
	)
	ukNameMap := make(map[string]struct{})
	for _, ukColName := range tblConfig.UniqueKeyColumnNames {
		ukNameMap[ukColName] = struct{}{}
		quotedUKColNames = append(quotedUKColNames, fmt.Sprintf("`%s`", ukColName))
	}
	for _, colDef := range tblConfig.Columns {
		if _, ok := ukNameMap[colDef.ColumnName]; !ok {
			nonUKColNames = append(nonUKColNames, fmt.Sprintf("`%s`", colDef.ColumnName))
		}
	}

	setExpr := fmt.Sprintf("(%s)=.+", strings.Join(nonUKColNames, "|"))
	whereExpr := fmt.Sprintf("(%s)=.+", strings.Join(quotedUKColNames, "|"))
	mock.ExpectExec(
		fmt.Sprintf("^UPDATE `%s`\\.`%s` SET (%s, ){%d}%s WHERE (%s AND ){%d}%s",
			tblConfig.DatabaseName,
			tblConfig.TableName,
			setExpr,
			len(nonUKColNames)-1,
			setExpr,
			whereExpr,
			len(tblConfig.UniqueKeyColumnNames)-1,
			whereExpr,
		),
	).WillReturnResult(sqlmock.NewResult(0, 1))
}

func mockDeleteStep(mock sqlmock.Sqlmock, tblConfig *config.TableConfig) {
	var (
		quotedUKColNames []string
	)
	for _, ukColName := range tblConfig.UniqueKeyColumnNames {
		quotedUKColNames = append(quotedUKColNames, fmt.Sprintf("`%s`", ukColName))
	}
	whereExpr := fmt.Sprintf("(%s)=.+", strings.Join(quotedUKColNames, "|"))
	mock.ExpectExec(
		fmt.Sprintf("^DELETE FROM `%s`\\.`%s` WHERE (%s AND ){%d}%s",
			tblConfig.DatabaseName,
			tblConfig.TableName,
			whereExpr,
			len(tblConfig.UniqueKeyColumnNames)-1,
			whereExpr,
		),
	).WillReturnResult(sqlmock.NewResult(0, 1))
}

func (s *testWorkloadSimulatorSuite) TestSchemaChange() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	s.Require().Nil(err)
	tblConfig01 := config.NewTemplateTableConfigForTest()
	tblConfig01.TableName = "members01"
	tblConfig01.TableID = "members01"
	tblConfig02 := config.NewTemplateTableConfigForTest()
	tblConfig02.TableName = "members02"
	tblConfig02.TableID = "members02"
	theSimulator, err := NewWorkloadSimulatorImpl(
		map[string]*config.TableConfig{
			"members01": tblConfig01,
			"members02": tblConfig02,
		},
		`INSERT members01;
		INSERT members02;
		UPDATE members01;
		UPDATE members02;
		DELETE members01;
		DELETE members02;
		`,
	)
	assert.Nil(s.T(), err)
	mcp01, err := newTestMCP(128, tblConfig01)
	s.Require().Nil(err)
	mcp02, err := newTestMCP(128, tblConfig02)
	s.Require().Nil(err)
	theMCPMap := map[string]*mcp.ModificationCandidatePool{
		tblConfig01.TableID: mcp01,
		tblConfig02.TableID: mcp02,
	}
	for i := 0; i < 50; i++ {
		mock.ExpectBegin()
		mockInsertStep(mock, tblConfig01)
		mockInsertStep(mock, tblConfig02)
		mockUpdateStep(mock, tblConfig01)
		mockUpdateStep(mock, tblConfig02)
		mockDeleteStep(mock, tblConfig01)
		mockDeleteStep(mock, tblConfig02)
		mock.ExpectCommit()
		err = theSimulator.SimulateTrx(ctx, db, theMCPMap)
		s.Require().Nil(err)
	}
	tblConfig02New := &config.TableConfig{
		TableID:      tblConfig02.TableID,
		DatabaseName: tblConfig02.DatabaseName,
		TableName:    tblConfig02.TableName,
		Columns: append(tblConfig02.Columns, &config.ColumnDefinition{
			ColumnName: "newcol",
			DataType:   "varchar",
		}),
		UniqueKeyColumnNames: []string{"name", "team_id"},
	}
	theSimulator.SetTableConfig("members02", tblConfig02New)
	mcp02New, err := newTestMCP(128, tblConfig02New)
	s.Require().Nil(err)
	theMCPMap[tblConfig02New.TableID] = mcp02New
	for i := 0; i < 50; i++ {
		mock.ExpectBegin()
		mockInsertStep(mock, tblConfig01)
		mockInsertStep(mock, tblConfig02New)
		mockUpdateStep(mock, tblConfig01)
		mockUpdateStep(mock, tblConfig02New)
		mockDeleteStep(mock, tblConfig01)
		mockDeleteStep(mock, tblConfig02New)
		mock.ExpectCommit()
		err = theSimulator.SimulateTrx(ctx, db, theMCPMap)
		s.Require().Nil(err)
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
				"members": config.NewTemplateTableConfigForTest(),
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
