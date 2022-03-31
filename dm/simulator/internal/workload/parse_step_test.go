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
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/parser"
)

type testParseStepSuite struct {
	suite.Suite
	tableConfigs map[string]*config.TableConfig
}

func (s *testParseStepSuite) SetupSuite() {
	s.Require().Nil(log.InitLogger(&log.Config{}))
	s.tableConfigs = map[string]*config.TableConfig{
		"games.members": {
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
		},
		"tbl2": {
			DatabaseName: "games",
			TableName:    "dummy",
			Columns: []*config.ColumnDefinition{
				{
					ColumnName: "id",
					DataType:   "int",
				},
				{
					ColumnName: "val",
					DataType:   "varchar",
				},
			},
			UniqueKeyColumnNames: []string{"id"},
		},
	}
}

func (s *testParseStepSuite) TestBasic() {
	testCode := `
REPEAT 2 (
  @testRow = INSERT games.members;
  REPEAT 3 (
      UPDATE @testRow;
  );
  DELETE @testRow;
);
REPEAT 3 (
  RANDOM-DML tbl2;
);`
	input := antlr.NewInputStream(testCode)
	lexer := parser.NewWorkloadLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewWorkloadParser(stream)
	el := NewParseStepsErrorListener(
		antlr.NewDiagnosticErrorListener(true),
	)
	p.AddErrorListener(el)
	p.BuildParseTrees = true
	tree := p.WorkloadSteps()
	sl := NewParseStepsListener(s.tableConfigs)
	antlr.ParseTreeWalkerDefault.Walk(sl, tree)
	s.Require().Nil(el.Err())
	for _, step := range sl.totalSteps {
		s.T().Logf("%v\n", step)
	}
}

func (s *testParseStepSuite) TestSyntaxError() {
	testCode := `
REPEAT 2 (
  testRow = INSERT games.members;
)
);`
	input := antlr.NewInputStream(testCode)
	lexer := parser.NewWorkloadLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewWorkloadParser(stream)
	el := NewParseStepsErrorListener(
		antlr.NewDiagnosticErrorListener(true),
	)
	p.AddErrorListener(el)
	p.BuildParseTrees = true
	tree := p.WorkloadSteps()
	sl := NewParseStepsListener(s.tableConfigs)
	antlr.ParseTreeWalkerDefault.Walk(sl, tree)
	s.Require().NotNil(el.Err(), "should have syntax error")
	s.T().Logf("Got syntax error: %v\n", el.Err())
}

func TestParseStepSuite(t *testing.T) {
	suite.Run(t, &testParseStepSuite{})
}
