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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
)

type testDMLSuite struct {
	suite.Suite
}

func (s *testDMLSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
}

func (s *testDMLSuite) TestDMLBasic() {
	var (
		err error
		sql string
		uk  *UniqueKey
	)
	g := NewDMLSQLGenerator()
	mcp := NewModificationCandidatePool()
	mcp.PreparePool()
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

func (s *testDMLSuite) TestDMLAbnormalUK() {
	var (
		sql string
		err error
		uk  *UniqueKey
	)
	g := NewDMLSQLGenerator()
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

func TestDMLSuite(t *testing.T) {
	suite.Run(t, &testDMLSuite{})
}
