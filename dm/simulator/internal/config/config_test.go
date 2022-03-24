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

package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testConfigSuite struct {
	suite.Suite
}

func (s *testConfigSuite) TestTableConfigDeepEqual() {
	var nilCfg *TableConfig
	cfg1 := NewTemplateTableConfigForTest()
	cfg1.UniqueKeyColumnNames = []string{"name", "team_id"}
	cfg2 := NewTemplateTableConfigForTest()
	cfg2.UniqueKeyColumnNames = []string{"name", "team_id"}
	s.True(cfg1.IsDeepEqual(cfg2))
	s.True(cfg2.IsDeepEqual(cfg1))

	s.False(nilCfg.IsDeepEqual(cfg1))
	s.False(cfg1.IsDeepEqual(nilCfg))

	clonedCfg1 := cfg1.SortedClone()
	s.True(cfg1.IsDeepEqual(clonedCfg1))
	s.True(clonedCfg1.IsDeepEqual(cfg1))

	clonedCfg1.Columns[0].DataType = "datetime"
	s.False(cfg1.IsDeepEqual(clonedCfg1))
	s.False(clonedCfg1.IsDeepEqual(cfg1))

	curTableID := cfg1.TableID
	cfg1.TableID = "aaaa"
	s.False(cfg1.IsDeepEqual(cfg2))
	s.False(cfg2.IsDeepEqual(cfg1))
	cfg1.TableID = curTableID

	curColDefs := cfg1.Columns
	cfg1.Columns = []*ColumnDefinition{
		cfg2.Columns[3],
		cfg2.Columns[2],
		cfg2.Columns[1],
		cfg2.Columns[0],
	}
	s.True(cfg1.IsDeepEqual(cfg2))
	s.True(cfg2.IsDeepEqual(cfg1))
	cfg1.Columns = curColDefs

	curColDefs = cfg1.Columns
	cfg1.Columns = append(cfg1.Columns, &ColumnDefinition{
		ColumnName: "newcol",
		DataType:   "int",
	})
	s.False(cfg1.IsDeepEqual(cfg2))
	s.False(cfg2.IsDeepEqual(cfg1))
	cfg1.Columns = curColDefs

	curColDef0 := cfg1.Columns[0]
	cfg1.Columns[0] = &ColumnDefinition{
		ColumnName: "newcol",
		DataType:   "int",
	}
	s.False(cfg1.IsDeepEqual(cfg2))
	s.False(cfg2.IsDeepEqual(cfg1))
	cfg1.Columns[0] = curColDef0

	curUKCols := cfg1.UniqueKeyColumnNames
	cfg1.UniqueKeyColumnNames = []string{
		cfg1.UniqueKeyColumnNames[1],
		cfg1.UniqueKeyColumnNames[0],
	}
	s.False(cfg1.IsDeepEqual(cfg2))
	s.False(cfg2.IsDeepEqual(cfg1))
	cfg1.UniqueKeyColumnNames = curUKCols

	curUKCols = cfg1.UniqueKeyColumnNames
	cfg1.UniqueKeyColumnNames = []string{"id"}
	s.False(cfg1.IsDeepEqual(cfg2))
	s.False(cfg2.IsDeepEqual(cfg1))
	cfg1.UniqueKeyColumnNames = curUKCols
}

func (s *testConfigSuite) TestUtils() {
	cfg := NewTemplateTableConfigForTest()
	colDefMap := GenerateColumnDefinitionsMap(cfg.Columns)
	for _, colDef := range cfg.Columns {
		colDefInMap, ok := colDefMap[colDef.ColumnName]
		s.Require().True(ok)
		s.Require().True(reflect.DeepEqual(colDef, colDefInMap))
	}
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, &testConfigSuite{})
}
