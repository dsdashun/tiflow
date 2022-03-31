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
	"reflect"
	"sort"
	"testing"

	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/stretchr/testify/suite"
)

type testMockWorkloadSuite struct {
	suite.Suite
}

func (s *testMockWorkloadSuite) TestPackTblConfigMapToString() {
	cfg1 := config.NewTemplateTableConfigForTest()
	cfg1.TableID = "tbl01"
	cfg1.TableName = "members01"
	cfg2 := config.NewTemplateTableConfigForTest()
	cfg2.TableID = "tbl02"
	cfg2.TableName = "members02"
	cfgMap := map[string]*config.TableConfig{
		cfg2.TableID: cfg2,
		cfg1.TableID: cfg1,
	}
	resultStr, err := packToString(cfgMap)
	s.Require().Nil(err)
	s.T().Log(resultStr)
	cfg1.Columns = []*config.ColumnDefinition{
		cfg1.Columns[3],
		cfg1.Columns[2],
		cfg1.Columns[1],
		cfg1.Columns[0],
	}
	resultStr2, err := packToString(cfgMap)
	s.Require().Nil(err)
	s.T().Log(resultStr2)
	s.Equal(resultStr, resultStr2)
}

func (s *testMockWorkloadSuite) TestBasic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg1 := config.NewTemplateTableConfigForTest()
	cfg1.TableID = "tbl01"
	cfg1.TableName = "members01"
	cfg2 := config.NewTemplateTableConfigForTest()
	cfg2.TableID = "tbl02"
	cfg2.TableName = "members02"
	ws, err := NewMockWorkload(map[string]*config.TableConfig{
		cfg2.TableID: cfg2,
		cfg1.TableID: cfg1,
	})
	s.Require().Nil(err)
	schemaSig, err := packToString(ws.tblConfigMap)
	s.Require().Nil(err)
	s.Require().Equal(schemaSig, ws.GetCurrentSchemaSignature())

	involvedTbls := ws.GetInvolvedTables()
	sort.Strings(involvedTbls)
	s.Require().True(reflect.DeepEqual(involvedTbls, []string{"tbl01", "tbl02"}))

	s.True(ws.DoesInvolveTable("tbl01"))
	s.True(ws.DoesInvolveTable("tbl02"))
	s.False(ws.DoesInvolveTable("tbl03"))

	repeatCnt := 100
	for i := 0; i < repeatCnt; i++ {
		s.Require().Nil(ws.SimulateTrx(ctx, nil, nil))
	}
	s.Equal(uint64(repeatCnt), ws.TotalExecuted(schemaSig))

	newCfg2 := cfg2.SortedClone()
	newCfg2.Columns = append(newCfg2.Columns, &config.ColumnDefinition{
		ColumnName: "dummycol",
		DataType:   "varchar",
	})
	s.Require().Nil(ws.SetTableConfig("tbl02", newCfg2))
	newSchemaSig, err := packToString(map[string]*config.TableConfig{
		cfg2.TableID: newCfg2,
		cfg1.TableID: cfg1,
	})
	s.Require().Nil(err)
	s.Require().NotEqual(schemaSig, newSchemaSig)
	s.Require().Equal(newSchemaSig, ws.GetCurrentSchemaSignature())
	for i := 0; i < repeatCnt; i++ {
		s.Require().Nil(ws.SimulateTrx(ctx, nil, nil))
	}
	s.Equal(uint64(repeatCnt), ws.TotalExecuted(newSchemaSig))

	ws.totalExecutedMap[newSchemaSig].Store(0) // reset the count
	if ws.IsEnabled() {
		ws.Disable()
	}
	for i := 0; i < repeatCnt; i++ {
		s.Require().Nil(ws.SimulateTrx(ctx, nil, nil))
	}
	s.Equal(uint64(0), ws.TotalExecuted(newSchemaSig))
	ws.Enable()
	for i := 0; i < repeatCnt; i++ {
		s.Require().Nil(ws.SimulateTrx(ctx, nil, nil))
	}
	s.Equal(uint64(repeatCnt), ws.TotalExecuted(newSchemaSig))
}

func TestMockWorkloadSuite(t *testing.T) {
	suite.Run(t, &testMockWorkloadSuite{})
}
