package core

import (
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
	s.tableConfigs = map[string]*config.TableConfig{
		"games.members": &config.TableConfig{
			DatabaseName: "games",
			TableName:    "members",
			Columns: []*config.ColumnDefinition{
				&config.ColumnDefinition{
					ColumnName: "id",
					DataType:   "int",
					DataLen:    11,
				},
				&config.ColumnDefinition{
					ColumnName: "name",
					DataType:   "varchar",
					DataLen:    255,
				},
				&config.ColumnDefinition{
					ColumnName: "age",
					DataType:   "int",
					DataLen:    11,
				},
				&config.ColumnDefinition{
					ColumnName: "team_id",
					DataType:   "int",
					DataLen:    11,
				},
			},
			UniqueKeyColumnNames: []string{"id"},
		},
		"tbl2": &config.TableConfig{
			DatabaseName: "games",
			TableName:    "dummy",
			Columns: []*config.ColumnDefinition{
				&config.ColumnDefinition{
					ColumnName: "id",
					DataType:   "int",
					DataLen:    11,
				},
				&config.ColumnDefinition{
					ColumnName: "val",
					DataType:   "varchar",
					DataLen:    255,
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
	assert.Nil(s.T(), el.Err())
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
	assert.NotNil(s.T(), el.Err(), "should have syntax error")
	s.T().Logf("Got syntax error: %v\n", el.Err())
}

func TestParseStepSuite(t *testing.T) {
	suite.Run(t, &testParseStepSuite{})
}
