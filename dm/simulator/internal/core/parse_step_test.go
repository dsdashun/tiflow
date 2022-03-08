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

type testWorkloadErrorListener struct {
	baseErrorListener antlr.ErrorListener
	t                 *testing.T
}

func NewTestWorkloadErrorListener(el antlr.ErrorListener, t *testing.T) *testWorkloadErrorListener {
	return &testWorkloadErrorListener{
		baseErrorListener: el,
		t:                 t,
	}
}

func (el *testWorkloadErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	el.baseErrorListener.SyntaxError(recognizer, offendingSymbol, line, column, msg, e)
	assert.Fail(el.t, "parsing code error", "syntax error")
}

func (el *testWorkloadErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	el.baseErrorListener.ReportAmbiguity(recognizer, dfa, startIndex, stopIndex, exact, ambigAlts, configs)
	assert.Fail(el.t, "parsing code error", "report ambiguity")
}

func (el *testWorkloadErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	el.baseErrorListener.ReportAttemptingFullContext(recognizer, dfa, startIndex, stopIndex, conflictingAlts, configs)
	assert.Fail(el.t, "parsing code error", "report attemping full context")
}

func (el *testWorkloadErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs antlr.ATNConfigSet) {
	el.baseErrorListener.ReportContextSensitivity(recognizer, dfa, startIndex, stopIndex, prediction, configs)
	assert.Fail(el.t, "parsing code error", "report context sensitivity")
}

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
	p.AddErrorListener(
		NewTestWorkloadErrorListener(
			antlr.NewDiagnosticErrorListener(true),
			s.T(),
		),
	)
	p.BuildParseTrees = true
	tree := p.WorkloadSteps()
	sl := NewParseStepsListener(s.tableConfigs)
	antlr.ParseTreeWalkerDefault.Walk(sl, tree)
	for _, step := range sl.totalSteps {
		s.T().Logf("%v\n", step)
	}
}

func TestParseStepSuite(t *testing.T) {
	suite.Run(t, &testParseStepSuite{})
}
