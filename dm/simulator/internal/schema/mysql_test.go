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

package schema

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
)

type testMySQLSchemaGenSuite struct {
	suite.Suite
}

func (s *testMySQLSchemaGenSuite) SetupSuite() {
	s.Require().Nil(log.InitLogger(&log.Config{}))
}

func (s *testMySQLSchemaGenSuite) TestGetTableDefinitionBasic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// TODO: change to mock SQL later
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", "root", "guanliyuanmima", "127.0.0.1", 13306))
	s.Require().Nil(err)
	schemaGetter := NewMySQLSchemaGetter(db, "games", "members")
	colDefs, err := schemaGetter.GetColumnDefinitions(ctx)
	s.Require().Nil(err)
	for _, col := range colDefs {
		s.T().Logf("column def: %+v\n", col)
	}
}

func (s *testMySQLSchemaGenSuite) TestGetUniqueKeyColumnsBasic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// TODO: change to mock SQL later
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", "root", "guanliyuanmima", "127.0.0.1", 13306))
	s.Require().Nil(err)
	schemaGetter := NewMySQLSchemaGetter(db, "games", "members")
	ukCols, err := schemaGetter.GetUniqueKeyColumns(ctx)
	s.Require().Nil(err)
	s.T().Logf("uk columns: %+v\n", ukCols)
}

func TestMySQLSchemaGenSuite(t *testing.T) {
	suite.Run(t, &testMySQLSchemaGenSuite{})
}
