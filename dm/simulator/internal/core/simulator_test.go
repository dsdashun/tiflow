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

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/chaos-mesh/go-sqlsmith/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type testSimulatorSuite struct {
	suite.Suite
	tableInfo *types.Table
	ukColumns map[string]*types.Column
}

func (s *testSimulatorSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
	s.tableInfo = &types.Table{
		DB:    "games",
		Table: "members",
		Type:  "BASE TABLE",
		Columns: map[string]*types.Column{
			"id": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "id",
				DataType: "int",
				DataLen:  11,
			},
			"name": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "name",
				DataType: "varchar",
				DataLen:  255,
			},
			"age": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "age",
				DataType: "int",
				DataLen:  11,
			},
			"team_id": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "team_id",
				DataType: "int",
				DataLen:  11,
			},
		},
	}
	s.ukColumns = map[string]*types.Column{
		"id": s.tableInfo.Columns["id"],
	}
}

func (s *testSimulatorSuite) TestSimulatorBasic() {
	db, err := sql.Open("mysql", "root:guanliyuanmima@tcp(127.0.0.1:13306)/games")
	// db, err := sql.Open("mysql", "root:@tcp(rms-staging.pingcap.net:31469)/games")
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(s.tableInfo, s.ukColumns)
	theSimulator := NewSimulatorImpl(db, sqlGen)
	err = theSimulator.prepareData(context.Background(), 128)
	assert.Nil(s.T(), err)
	err = theSimulator.loadMCP(context.Background())
	assert.Nil(s.T(), err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	theSimulator.DoSimulation(ctx)
	s.T().Logf("total executed trx: %d\n", theSimulator.totalExecutedTrx)
}

func (s *testSimulatorSuite) TestSingleSimulation() {
	var (
		err error
		sql string
		uk  *sqlgen.UniqueKey
	)
	prepareDataRecord := 128
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(s.tableInfo, s.ukColumns)
	theSimulator := NewSimulatorImpl(db, sqlGen)
	mock.ExpectBegin()
	sql, err = theSimulator.sqlGen.GenTruncateTable()
	if err != nil {
		s.T().Fatalf("generate truncate table error: %v\n", err)
	}
	mock.ExpectExec(sql).WillReturnResult(sqlmock.NewResult(0, int64(prepareDataRecord)))
	expectRows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < prepareDataRecord; i++ {
		mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
		expectRows.AddRow(i)
	}
	mock.ExpectCommit()
	err = theSimulator.prepareData(context.Background(), prepareDataRecord)
	assert.Nil(s.T(), err)

	sql, _, err = theSimulator.sqlGen.GenLoadUniqueKeySQL()
	if err != nil {
		s.T().Fatalf("generate truncate table error: %v\n", err)
	}
	mock.ExpectQuery(sql).WillReturnRows(expectRows)
	err = theSimulator.loadMCP(context.Background())
	assert.Nil(s.T(), err)
	mock.ExpectBegin()
	tx, err := theSimulator.db.BeginTx(ctx, nil)
	assert.Nil(s.T(), err)
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	uk, err = theSimulator.simulateInsert(ctx, tx)
	assert.Nil(s.T(), err)
	s.T().Logf("new UK: %v\n", uk)
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theSimulator.simulateUpdate(ctx, tx)
	assert.Nil(s.T(), err)
	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theSimulator.simulateDelete(ctx, tx)
	assert.Nil(s.T(), err)
	mock.ExpectCommit()
	err = tx.Commit()
	assert.Nil(s.T(), err)
}

func TestSimulatorSuite(t *testing.T) {
	suite.Run(t, &testSimulatorSuite{})
}
