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
	"reflect"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
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
	db, mock, err := sqlmock.New()
	s.Require().Nil(err)
	tblConfig := config.NewTemplateTableConfigForTest()
	tabDefRows := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE"})
	for _, colDef := range tblConfig.Columns {
		tabDefRows.AddRow(colDef.ColumnName, colDef.DataType)
	}
	dbName := "games"
	tblName := "members"
	mock.ExpectQuery(regexp.QuoteMeta(sqlGetColumnDefinitions)).WithArgs(dbName, tblName).WillReturnRows(tabDefRows)
	schemaGetter := NewMySQLSchemaGetter(db)
	colDefs, err := schemaGetter.GetColumnDefinitions(ctx, dbName, tblName)
	s.Require().Nil(err)
	for i, colDef := range colDefs {
		s.T().Logf("column def: %+v\n", colDef)
		s.True(reflect.DeepEqual(tblConfig.Columns[i], colDef))
	}
}

func (s *testMySQLSchemaGenSuite) TestGetUniqueKeyColumnsBasic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	s.Require().Nil(err)
	idxInfoRows := sqlmock.NewRows([]string{
		"Table",
		"Non_unique",
		"Key_name",
		"Seq_in_index",
		"Column_name",
		"Collation",
		"Cardinality",
		"Sub_part",
		"Packed",
		"Null",
		"Index_type",
		"Comment",
		"Index_comment",
	})
	dbName := "games"
	tblName := "members"
	idxInfoRows.AddRow(tblName, 0, "PRIMARY", 1, "id", "A", 3571, sql.NullString{}, sql.NullString{}, "", "BTREE", "", "")
	idxInfoRows.AddRow(tblName, 0, "name_team_id", 2, "team_id", "A", 3531, sql.NullString{}, sql.NullString{}, "YES", "BTREE", "", "")
	idxInfoRows.AddRow(tblName, 0, "name_team_id", 1, "name", "A", 3463, sql.NullString{}, sql.NullString{}, "YES", "BTREE", "", "")
	mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf(sqlGetIndex, dbName, tblName))).WillReturnRows(idxInfoRows)
	schemaGetter := NewMySQLSchemaGetter(db)
	ukCols, err := schemaGetter.GetUniqueKeyColumns(ctx, dbName, tblName)
	s.Require().Nil(err)
	s.T().Logf("uk columns: %+v\n", ukCols)
	s.True(reflect.DeepEqual(ukCols, []string{"id"}))
}

func TestMySQLSchemaGenSuite(t *testing.T) {
	suite.Run(t, &testMySQLSchemaGenSuite{})
}
