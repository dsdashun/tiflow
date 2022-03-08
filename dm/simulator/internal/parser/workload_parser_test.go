package parser

import (
	"fmt"
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
)

type workloadListener struct {
	*BaseWorkloadListener
}

func NewWorkloadListener() *workloadListener {
	return new(workloadListener)
}

func (l *workloadListener) EnterWorkloadStep(c *WorkloadStepContext) {
	fmt.Printf("Workload step: %s\n", c.ToStringTree(nil, c.GetParser()))
}

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

type testParserSuite struct {
	suite.Suite
}

func (s *testParserSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
}

func (s *testParserSuite) TestBasic() {
	testCode := `
REPEAT 2 (
  @a = INSERT abcd;
  REPEAT 3 (
      UPDATE @a;
  );
  DELETE @a;
);
REPEAT 3 (
  RANDOM-DML tbl2;
);`
	input := antlr.NewInputStream(testCode)
	lexer := NewWorkloadLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := NewWorkloadParser(stream)
	p.AddErrorListener(
		NewTestWorkloadErrorListener(
			antlr.NewDiagnosticErrorListener(true),
			s.T(),
		),
	)
	p.BuildParseTrees = true
	tree := p.WorkloadSteps()
	antlr.ParseTreeWalkerDefault.Walk(NewWorkloadListener(), tree)
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, &testParserSuite{})
}
