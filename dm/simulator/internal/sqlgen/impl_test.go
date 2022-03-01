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
	"testing"

	"github.com/chaos-mesh/go-sqlsmith/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
)

type testSQLGenImplSuite struct {
	suite.Suite
	tableInfo *types.Table
	ukColumns map[string]*types.Column
}

func (s *testSQLGenImplSuite) SetupSuite() {
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

func (s *testSQLGenImplSuite) TestDMLBasic() {
	var (
		err error
		sql string
		uk  *UniqueKey
	)
	g := NewSQLGeneratorImpl(s.tableInfo, s.ukColumns)
	mcp := NewModificationCandidatePool()
	mcp.PreparePool()

	sql, _, err = g.GenLoadUniqueKeySQL()
	assert.Nil(s.T(), err)
	s.T().Logf("Generated SELECT SQL: %s\n", sql)

	sql, err = g.GenTruncateTable()
	assert.Nil(s.T(), err)
	s.T().Logf("Generated Truncate Table SQL: %s\n", sql)

	var ukIter UniqueKeyIterator = mcp
	for i := 0; i < 10; i++ {
		uk = ukIter.NextUK()
		sql, err = g.GenUpdateRow(uk)
		assert.Nil(s.T(), err)
		s.T().Logf("Generated SQL: %s\n", sql)
		sql, uk, err = g.GenInsertRow()
		assert.Nil(s.T(), err)
		s.T().Logf("Generated SQL: %s\n; Unique key: %v\n", sql, uk)
		uk = ukIter.NextUK()
		sql, err = g.GenDeleteRow(uk)
		assert.Nil(s.T(), err)
		s.T().Logf("Generated SQL: %s\n; Unique key: %v\n", sql, uk)
	}
}

func (s *testSQLGenImplSuite) TestDMLAbnormalUK() {
	var (
		sql string
		err error
		uk  *UniqueKey
	)
	g := NewSQLGeneratorImpl(s.tableInfo, s.ukColumns)
	uk = &UniqueKey{
		RowID: -1,
		Value: map[string]interface{}{
			"abcdefg": 123,
		},
	}
	_, err = g.GenUpdateRow(uk)
	assert.NotNil(s.T(), err)
	_, err = g.GenDeleteRow(uk)
	assert.NotNil(s.T(), err)

	uk = &UniqueKey{
		RowID: -1,
		Value: map[string]interface{}{
			"id":      123,
			"abcdefg": 321,
		},
	}
	sql, err = g.GenUpdateRow(uk)
	assert.Nil(s.T(), err)
	s.T().Logf("Generated SQL: %s\n", sql)
	sql, err = g.GenDeleteRow(uk)
	assert.Nil(s.T(), err)
	s.T().Logf("Generated SQL: %s\n", sql)
}

func TestSQLGenImplSuite(t *testing.T) {
	suite.Run(t, &testSQLGenImplSuite{})
}