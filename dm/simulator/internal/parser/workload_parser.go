// Code generated from Workload.g4 by ANTLR 4.9.3. DO NOT EDIT.

package parser // Workload

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 18, 94, 4,
	2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7, 4,
	8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4, 13, 9,
	13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 3, 2, 6, 2, 34, 10, 2, 13,
	2, 14, 2, 35, 3, 3, 3, 3, 3, 4, 3, 4, 5, 4, 42, 10, 4, 3, 4, 3, 4, 3, 5,
	3, 5, 3, 5, 3, 5, 6, 5, 50, 10, 5, 13, 5, 14, 5, 51, 3, 5, 3, 5, 3, 6,
	3, 6, 3, 6, 3, 6, 5, 6, 60, 10, 6, 3, 7, 3, 7, 5, 7, 64, 10, 7, 3, 8, 3,
	8, 3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 3, 10, 3, 10, 5, 10, 75, 10, 10, 3, 11,
	3, 11, 3, 11, 3, 11, 3, 12, 3, 12, 3, 12, 3, 13, 3, 13, 3, 14, 3, 14, 3,
	14, 3, 15, 3, 15, 3, 15, 3, 16, 3, 16, 3, 16, 2, 2, 17, 2, 4, 6, 8, 10,
	12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 2, 3, 3, 2, 12, 13, 2, 86, 2, 33,
	3, 2, 2, 2, 4, 37, 3, 2, 2, 2, 6, 41, 3, 2, 2, 2, 8, 45, 3, 2, 2, 2, 10,
	59, 3, 2, 2, 2, 12, 63, 3, 2, 2, 2, 14, 65, 3, 2, 2, 2, 16, 69, 3, 2, 2,
	2, 18, 74, 3, 2, 2, 2, 20, 76, 3, 2, 2, 2, 22, 80, 3, 2, 2, 2, 24, 83,
	3, 2, 2, 2, 26, 85, 3, 2, 2, 2, 28, 88, 3, 2, 2, 2, 30, 91, 3, 2, 2, 2,
	32, 34, 5, 4, 3, 2, 33, 32, 3, 2, 2, 2, 34, 35, 3, 2, 2, 2, 35, 33, 3,
	2, 2, 2, 35, 36, 3, 2, 2, 2, 36, 3, 3, 2, 2, 2, 37, 38, 5, 6, 4, 2, 38,
	5, 3, 2, 2, 2, 39, 42, 5, 8, 5, 2, 40, 42, 5, 10, 6, 2, 41, 39, 3, 2, 2,
	2, 41, 40, 3, 2, 2, 2, 42, 43, 3, 2, 2, 2, 43, 44, 7, 3, 2, 2, 44, 7, 3,
	2, 2, 2, 45, 46, 7, 4, 2, 2, 46, 47, 7, 16, 2, 2, 47, 49, 7, 5, 2, 2, 48,
	50, 5, 6, 4, 2, 49, 48, 3, 2, 2, 2, 50, 51, 3, 2, 2, 2, 51, 49, 3, 2, 2,
	2, 51, 52, 3, 2, 2, 2, 52, 53, 3, 2, 2, 2, 53, 54, 7, 6, 2, 2, 54, 9, 3,
	2, 2, 2, 55, 60, 5, 12, 7, 2, 56, 60, 5, 18, 10, 2, 57, 60, 5, 24, 13,
	2, 58, 60, 5, 28, 15, 2, 59, 55, 3, 2, 2, 2, 59, 56, 3, 2, 2, 2, 59, 57,
	3, 2, 2, 2, 59, 58, 3, 2, 2, 2, 60, 11, 3, 2, 2, 2, 61, 64, 5, 16, 9, 2,
	62, 64, 5, 14, 8, 2, 63, 61, 3, 2, 2, 2, 63, 62, 3, 2, 2, 2, 64, 13, 3,
	2, 2, 2, 65, 66, 7, 13, 2, 2, 66, 67, 7, 7, 2, 2, 67, 68, 5, 16, 9, 2,
	68, 15, 3, 2, 2, 2, 69, 70, 7, 8, 2, 2, 70, 71, 7, 12, 2, 2, 71, 17, 3,
	2, 2, 2, 72, 75, 5, 22, 12, 2, 73, 75, 5, 20, 11, 2, 74, 72, 3, 2, 2, 2,
	74, 73, 3, 2, 2, 2, 75, 19, 3, 2, 2, 2, 76, 77, 7, 13, 2, 2, 77, 78, 7,
	7, 2, 2, 78, 79, 5, 22, 12, 2, 79, 21, 3, 2, 2, 2, 80, 81, 7, 9, 2, 2,
	81, 82, 5, 30, 16, 2, 82, 23, 3, 2, 2, 2, 83, 84, 5, 26, 14, 2, 84, 25,
	3, 2, 2, 2, 85, 86, 7, 10, 2, 2, 86, 87, 5, 30, 16, 2, 87, 27, 3, 2, 2,
	2, 88, 89, 7, 11, 2, 2, 89, 90, 7, 12, 2, 2, 90, 29, 3, 2, 2, 2, 91, 92,
	9, 2, 2, 2, 92, 31, 3, 2, 2, 2, 8, 35, 41, 51, 59, 63, 74,
}
var literalNames = []string{
	"", "';'", "'REPEAT'", "'('", "')'", "'='", "'INSERT'", "'UPDATE'", "'DELETE'",
	"'RANDOM-DML'",
}
var symbolicNames = []string{
	"", "", "", "", "", "", "", "", "", "", "TableName", "RowID", "UID", "ReverseQuoteID",
	"Int", "ID", "WS",
}

var ruleNames = []string{
	"workloadSteps", "workloadStep", "singleStep", "repeatStep", "dmlStep",
	"insertOp", "assignmentInsertOp", "simpleInsertOp", "updateOp", "assignmentUpdateOp",
	"simpleUpdateOp", "deleteOp", "simpleDeleteOp", "randomDmlOp", "tableOrRowRef",
}

type WorkloadParser struct {
	*antlr.BaseParser
}

// NewWorkloadParser produces a new parser instance for the optional input antlr.TokenStream.
//
// The *WorkloadParser instance produced may be reused by calling the SetInputStream method.
// The initial parser configuration is expensive to construct, and the object is not thread-safe;
// however, if used within a Golang sync.Pool, the construction cost amortizes well and the
// objects can be used in a thread-safe manner.
func NewWorkloadParser(input antlr.TokenStream) *WorkloadParser {
	this := new(WorkloadParser)
	deserializer := antlr.NewATNDeserializer(nil)
	deserializedATN := deserializer.DeserializeFromUInt16(parserATN)
	decisionToDFA := make([]*antlr.DFA, len(deserializedATN.DecisionToState))
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "Workload.g4"

	return this
}

// WorkloadParser tokens.
const (
	WorkloadParserEOF            = antlr.TokenEOF
	WorkloadParserT__0           = 1
	WorkloadParserT__1           = 2
	WorkloadParserT__2           = 3
	WorkloadParserT__3           = 4
	WorkloadParserT__4           = 5
	WorkloadParserT__5           = 6
	WorkloadParserT__6           = 7
	WorkloadParserT__7           = 8
	WorkloadParserT__8           = 9
	WorkloadParserTableName      = 10
	WorkloadParserRowID          = 11
	WorkloadParserUID            = 12
	WorkloadParserReverseQuoteID = 13
	WorkloadParserInt            = 14
	WorkloadParserID             = 15
	WorkloadParserWS             = 16
)

// WorkloadParser rules.
const (
	WorkloadParserRULE_workloadSteps      = 0
	WorkloadParserRULE_workloadStep       = 1
	WorkloadParserRULE_singleStep         = 2
	WorkloadParserRULE_repeatStep         = 3
	WorkloadParserRULE_dmlStep            = 4
	WorkloadParserRULE_insertOp           = 5
	WorkloadParserRULE_assignmentInsertOp = 6
	WorkloadParserRULE_simpleInsertOp     = 7
	WorkloadParserRULE_updateOp           = 8
	WorkloadParserRULE_assignmentUpdateOp = 9
	WorkloadParserRULE_simpleUpdateOp     = 10
	WorkloadParserRULE_deleteOp           = 11
	WorkloadParserRULE_simpleDeleteOp     = 12
	WorkloadParserRULE_randomDmlOp        = 13
	WorkloadParserRULE_tableOrRowRef      = 14
)

// IWorkloadStepsContext is an interface to support dynamic dispatch.
type IWorkloadStepsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsWorkloadStepsContext differentiates from other interfaces.
	IsWorkloadStepsContext()
}

type WorkloadStepsContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWorkloadStepsContext() *WorkloadStepsContext {
	var p = new(WorkloadStepsContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_workloadSteps
	return p
}

func (*WorkloadStepsContext) IsWorkloadStepsContext() {}

func NewWorkloadStepsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WorkloadStepsContext {
	var p = new(WorkloadStepsContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_workloadSteps

	return p
}

func (s *WorkloadStepsContext) GetParser() antlr.Parser { return s.parser }

func (s *WorkloadStepsContext) AllWorkloadStep() []IWorkloadStepContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IWorkloadStepContext)(nil)).Elem())
	var tst = make([]IWorkloadStepContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IWorkloadStepContext)
		}
	}

	return tst
}

func (s *WorkloadStepsContext) WorkloadStep(i int) IWorkloadStepContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWorkloadStepContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IWorkloadStepContext)
}

func (s *WorkloadStepsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WorkloadStepsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WorkloadStepsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterWorkloadSteps(s)
	}
}

func (s *WorkloadStepsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitWorkloadSteps(s)
	}
}

func (p *WorkloadParser) WorkloadSteps() (localctx IWorkloadStepsContext) {
	this := p
	_ = this

	localctx = NewWorkloadStepsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, WorkloadParserRULE_workloadSteps)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(31)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<WorkloadParserT__1)|(1<<WorkloadParserT__5)|(1<<WorkloadParserT__6)|(1<<WorkloadParserT__7)|(1<<WorkloadParserT__8)|(1<<WorkloadParserRowID))) != 0) {
		{
			p.SetState(30)
			p.WorkloadStep()
		}

		p.SetState(33)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}

	return localctx
}

// IWorkloadStepContext is an interface to support dynamic dispatch.
type IWorkloadStepContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsWorkloadStepContext differentiates from other interfaces.
	IsWorkloadStepContext()
}

type WorkloadStepContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWorkloadStepContext() *WorkloadStepContext {
	var p = new(WorkloadStepContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_workloadStep
	return p
}

func (*WorkloadStepContext) IsWorkloadStepContext() {}

func NewWorkloadStepContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WorkloadStepContext {
	var p = new(WorkloadStepContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_workloadStep

	return p
}

func (s *WorkloadStepContext) GetParser() antlr.Parser { return s.parser }

func (s *WorkloadStepContext) SingleStep() ISingleStepContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISingleStepContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISingleStepContext)
}

func (s *WorkloadStepContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WorkloadStepContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WorkloadStepContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterWorkloadStep(s)
	}
}

func (s *WorkloadStepContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitWorkloadStep(s)
	}
}

func (p *WorkloadParser) WorkloadStep() (localctx IWorkloadStepContext) {
	this := p
	_ = this

	localctx = NewWorkloadStepContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, WorkloadParserRULE_workloadStep)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(35)
		p.SingleStep()
	}

	return localctx
}

// ISingleStepContext is an interface to support dynamic dispatch.
type ISingleStepContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSingleStepContext differentiates from other interfaces.
	IsSingleStepContext()
}

type SingleStepContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySingleStepContext() *SingleStepContext {
	var p = new(SingleStepContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_singleStep
	return p
}

func (*SingleStepContext) IsSingleStepContext() {}

func NewSingleStepContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SingleStepContext {
	var p = new(SingleStepContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_singleStep

	return p
}

func (s *SingleStepContext) GetParser() antlr.Parser { return s.parser }

func (s *SingleStepContext) RepeatStep() IRepeatStepContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IRepeatStepContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IRepeatStepContext)
}

func (s *SingleStepContext) DmlStep() IDmlStepContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDmlStepContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDmlStepContext)
}

func (s *SingleStepContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SingleStepContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SingleStepContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterSingleStep(s)
	}
}

func (s *SingleStepContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitSingleStep(s)
	}
}

func (p *WorkloadParser) SingleStep() (localctx ISingleStepContext) {
	this := p
	_ = this

	localctx = NewSingleStepContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, WorkloadParserRULE_singleStep)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(39)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case WorkloadParserT__1:
		{
			p.SetState(37)
			p.RepeatStep()
		}

	case WorkloadParserT__5, WorkloadParserT__6, WorkloadParserT__7, WorkloadParserT__8, WorkloadParserRowID:
		{
			p.SetState(38)
			p.DmlStep()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}
	{
		p.SetState(41)
		p.Match(WorkloadParserT__0)
	}

	return localctx
}

// IRepeatStepContext is an interface to support dynamic dispatch.
type IRepeatStepContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsRepeatStepContext differentiates from other interfaces.
	IsRepeatStepContext()
}

type RepeatStepContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRepeatStepContext() *RepeatStepContext {
	var p = new(RepeatStepContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_repeatStep
	return p
}

func (*RepeatStepContext) IsRepeatStepContext() {}

func NewRepeatStepContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RepeatStepContext {
	var p = new(RepeatStepContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_repeatStep

	return p
}

func (s *RepeatStepContext) GetParser() antlr.Parser { return s.parser }

func (s *RepeatStepContext) Int() antlr.TerminalNode {
	return s.GetToken(WorkloadParserInt, 0)
}

func (s *RepeatStepContext) AllSingleStep() []ISingleStepContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISingleStepContext)(nil)).Elem())
	var tst = make([]ISingleStepContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISingleStepContext)
		}
	}

	return tst
}

func (s *RepeatStepContext) SingleStep(i int) ISingleStepContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISingleStepContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISingleStepContext)
}

func (s *RepeatStepContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RepeatStepContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RepeatStepContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterRepeatStep(s)
	}
}

func (s *RepeatStepContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitRepeatStep(s)
	}
}

func (p *WorkloadParser) RepeatStep() (localctx IRepeatStepContext) {
	this := p
	_ = this

	localctx = NewRepeatStepContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, WorkloadParserRULE_repeatStep)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(43)
		p.Match(WorkloadParserT__1)
	}
	{
		p.SetState(44)
		p.Match(WorkloadParserInt)
	}
	{
		p.SetState(45)
		p.Match(WorkloadParserT__2)
	}
	p.SetState(47)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<WorkloadParserT__1)|(1<<WorkloadParserT__5)|(1<<WorkloadParserT__6)|(1<<WorkloadParserT__7)|(1<<WorkloadParserT__8)|(1<<WorkloadParserRowID))) != 0) {
		{
			p.SetState(46)
			p.SingleStep()
		}

		p.SetState(49)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(51)
		p.Match(WorkloadParserT__3)
	}

	return localctx
}

// IDmlStepContext is an interface to support dynamic dispatch.
type IDmlStepContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsDmlStepContext differentiates from other interfaces.
	IsDmlStepContext()
}

type DmlStepContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDmlStepContext() *DmlStepContext {
	var p = new(DmlStepContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_dmlStep
	return p
}

func (*DmlStepContext) IsDmlStepContext() {}

func NewDmlStepContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DmlStepContext {
	var p = new(DmlStepContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_dmlStep

	return p
}

func (s *DmlStepContext) GetParser() antlr.Parser { return s.parser }

func (s *DmlStepContext) InsertOp() IInsertOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IInsertOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IInsertOpContext)
}

func (s *DmlStepContext) UpdateOp() IUpdateOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IUpdateOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IUpdateOpContext)
}

func (s *DmlStepContext) DeleteOp() IDeleteOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDeleteOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDeleteOpContext)
}

func (s *DmlStepContext) RandomDmlOp() IRandomDmlOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IRandomDmlOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IRandomDmlOpContext)
}

func (s *DmlStepContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DmlStepContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DmlStepContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterDmlStep(s)
	}
}

func (s *DmlStepContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitDmlStep(s)
	}
}

func (p *WorkloadParser) DmlStep() (localctx IDmlStepContext) {
	this := p
	_ = this

	localctx = NewDmlStepContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, WorkloadParserRULE_dmlStep)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(53)
			p.InsertOp()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(54)
			p.UpdateOp()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(55)
			p.DeleteOp()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(56)
			p.RandomDmlOp()
		}

	}

	return localctx
}

// IInsertOpContext is an interface to support dynamic dispatch.
type IInsertOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsInsertOpContext differentiates from other interfaces.
	IsInsertOpContext()
}

type InsertOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInsertOpContext() *InsertOpContext {
	var p = new(InsertOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_insertOp
	return p
}

func (*InsertOpContext) IsInsertOpContext() {}

func NewInsertOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InsertOpContext {
	var p = new(InsertOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_insertOp

	return p
}

func (s *InsertOpContext) GetParser() antlr.Parser { return s.parser }

func (s *InsertOpContext) SimpleInsertOp() ISimpleInsertOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISimpleInsertOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISimpleInsertOpContext)
}

func (s *InsertOpContext) AssignmentInsertOp() IAssignmentInsertOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IAssignmentInsertOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IAssignmentInsertOpContext)
}

func (s *InsertOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InsertOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InsertOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterInsertOp(s)
	}
}

func (s *InsertOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitInsertOp(s)
	}
}

func (p *WorkloadParser) InsertOp() (localctx IInsertOpContext) {
	this := p
	_ = this

	localctx = NewInsertOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, WorkloadParserRULE_insertOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(61)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case WorkloadParserT__5:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(59)
			p.SimpleInsertOp()
		}

	case WorkloadParserRowID:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(60)
			p.AssignmentInsertOp()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IAssignmentInsertOpContext is an interface to support dynamic dispatch.
type IAssignmentInsertOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsAssignmentInsertOpContext differentiates from other interfaces.
	IsAssignmentInsertOpContext()
}

type AssignmentInsertOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAssignmentInsertOpContext() *AssignmentInsertOpContext {
	var p = new(AssignmentInsertOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_assignmentInsertOp
	return p
}

func (*AssignmentInsertOpContext) IsAssignmentInsertOpContext() {}

func NewAssignmentInsertOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AssignmentInsertOpContext {
	var p = new(AssignmentInsertOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_assignmentInsertOp

	return p
}

func (s *AssignmentInsertOpContext) GetParser() antlr.Parser { return s.parser }

func (s *AssignmentInsertOpContext) RowID() antlr.TerminalNode {
	return s.GetToken(WorkloadParserRowID, 0)
}

func (s *AssignmentInsertOpContext) SimpleInsertOp() ISimpleInsertOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISimpleInsertOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISimpleInsertOpContext)
}

func (s *AssignmentInsertOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AssignmentInsertOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AssignmentInsertOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterAssignmentInsertOp(s)
	}
}

func (s *AssignmentInsertOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitAssignmentInsertOp(s)
	}
}

func (p *WorkloadParser) AssignmentInsertOp() (localctx IAssignmentInsertOpContext) {
	this := p
	_ = this

	localctx = NewAssignmentInsertOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, WorkloadParserRULE_assignmentInsertOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(63)
		p.Match(WorkloadParserRowID)
	}
	{
		p.SetState(64)
		p.Match(WorkloadParserT__4)
	}
	{
		p.SetState(65)
		p.SimpleInsertOp()
	}

	return localctx
}

// ISimpleInsertOpContext is an interface to support dynamic dispatch.
type ISimpleInsertOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSimpleInsertOpContext differentiates from other interfaces.
	IsSimpleInsertOpContext()
}

type SimpleInsertOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySimpleInsertOpContext() *SimpleInsertOpContext {
	var p = new(SimpleInsertOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_simpleInsertOp
	return p
}

func (*SimpleInsertOpContext) IsSimpleInsertOpContext() {}

func NewSimpleInsertOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SimpleInsertOpContext {
	var p = new(SimpleInsertOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_simpleInsertOp

	return p
}

func (s *SimpleInsertOpContext) GetParser() antlr.Parser { return s.parser }

func (s *SimpleInsertOpContext) TableName() antlr.TerminalNode {
	return s.GetToken(WorkloadParserTableName, 0)
}

func (s *SimpleInsertOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SimpleInsertOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SimpleInsertOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterSimpleInsertOp(s)
	}
}

func (s *SimpleInsertOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitSimpleInsertOp(s)
	}
}

func (p *WorkloadParser) SimpleInsertOp() (localctx ISimpleInsertOpContext) {
	this := p
	_ = this

	localctx = NewSimpleInsertOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, WorkloadParserRULE_simpleInsertOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(67)
		p.Match(WorkloadParserT__5)
	}
	{
		p.SetState(68)
		p.Match(WorkloadParserTableName)
	}

	return localctx
}

// IUpdateOpContext is an interface to support dynamic dispatch.
type IUpdateOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsUpdateOpContext differentiates from other interfaces.
	IsUpdateOpContext()
}

type UpdateOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUpdateOpContext() *UpdateOpContext {
	var p = new(UpdateOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_updateOp
	return p
}

func (*UpdateOpContext) IsUpdateOpContext() {}

func NewUpdateOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UpdateOpContext {
	var p = new(UpdateOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_updateOp

	return p
}

func (s *UpdateOpContext) GetParser() antlr.Parser { return s.parser }

func (s *UpdateOpContext) SimpleUpdateOp() ISimpleUpdateOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISimpleUpdateOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISimpleUpdateOpContext)
}

func (s *UpdateOpContext) AssignmentUpdateOp() IAssignmentUpdateOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IAssignmentUpdateOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IAssignmentUpdateOpContext)
}

func (s *UpdateOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UpdateOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UpdateOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterUpdateOp(s)
	}
}

func (s *UpdateOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitUpdateOp(s)
	}
}

func (p *WorkloadParser) UpdateOp() (localctx IUpdateOpContext) {
	this := p
	_ = this

	localctx = NewUpdateOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, WorkloadParserRULE_updateOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(72)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case WorkloadParserT__6:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(70)
			p.SimpleUpdateOp()
		}

	case WorkloadParserRowID:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(71)
			p.AssignmentUpdateOp()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IAssignmentUpdateOpContext is an interface to support dynamic dispatch.
type IAssignmentUpdateOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsAssignmentUpdateOpContext differentiates from other interfaces.
	IsAssignmentUpdateOpContext()
}

type AssignmentUpdateOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAssignmentUpdateOpContext() *AssignmentUpdateOpContext {
	var p = new(AssignmentUpdateOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_assignmentUpdateOp
	return p
}

func (*AssignmentUpdateOpContext) IsAssignmentUpdateOpContext() {}

func NewAssignmentUpdateOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AssignmentUpdateOpContext {
	var p = new(AssignmentUpdateOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_assignmentUpdateOp

	return p
}

func (s *AssignmentUpdateOpContext) GetParser() antlr.Parser { return s.parser }

func (s *AssignmentUpdateOpContext) RowID() antlr.TerminalNode {
	return s.GetToken(WorkloadParserRowID, 0)
}

func (s *AssignmentUpdateOpContext) SimpleUpdateOp() ISimpleUpdateOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISimpleUpdateOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISimpleUpdateOpContext)
}

func (s *AssignmentUpdateOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AssignmentUpdateOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AssignmentUpdateOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterAssignmentUpdateOp(s)
	}
}

func (s *AssignmentUpdateOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitAssignmentUpdateOp(s)
	}
}

func (p *WorkloadParser) AssignmentUpdateOp() (localctx IAssignmentUpdateOpContext) {
	this := p
	_ = this

	localctx = NewAssignmentUpdateOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, WorkloadParserRULE_assignmentUpdateOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(74)
		p.Match(WorkloadParserRowID)
	}
	{
		p.SetState(75)
		p.Match(WorkloadParserT__4)
	}
	{
		p.SetState(76)
		p.SimpleUpdateOp()
	}

	return localctx
}

// ISimpleUpdateOpContext is an interface to support dynamic dispatch.
type ISimpleUpdateOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSimpleUpdateOpContext differentiates from other interfaces.
	IsSimpleUpdateOpContext()
}

type SimpleUpdateOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySimpleUpdateOpContext() *SimpleUpdateOpContext {
	var p = new(SimpleUpdateOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_simpleUpdateOp
	return p
}

func (*SimpleUpdateOpContext) IsSimpleUpdateOpContext() {}

func NewSimpleUpdateOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SimpleUpdateOpContext {
	var p = new(SimpleUpdateOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_simpleUpdateOp

	return p
}

func (s *SimpleUpdateOpContext) GetParser() antlr.Parser { return s.parser }

func (s *SimpleUpdateOpContext) TableOrRowRef() ITableOrRowRefContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITableOrRowRefContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITableOrRowRefContext)
}

func (s *SimpleUpdateOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SimpleUpdateOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SimpleUpdateOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterSimpleUpdateOp(s)
	}
}

func (s *SimpleUpdateOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitSimpleUpdateOp(s)
	}
}

func (p *WorkloadParser) SimpleUpdateOp() (localctx ISimpleUpdateOpContext) {
	this := p
	_ = this

	localctx = NewSimpleUpdateOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, WorkloadParserRULE_simpleUpdateOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(78)
		p.Match(WorkloadParserT__6)
	}
	{
		p.SetState(79)
		p.TableOrRowRef()
	}

	return localctx
}

// IDeleteOpContext is an interface to support dynamic dispatch.
type IDeleteOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsDeleteOpContext differentiates from other interfaces.
	IsDeleteOpContext()
}

type DeleteOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDeleteOpContext() *DeleteOpContext {
	var p = new(DeleteOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_deleteOp
	return p
}

func (*DeleteOpContext) IsDeleteOpContext() {}

func NewDeleteOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DeleteOpContext {
	var p = new(DeleteOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_deleteOp

	return p
}

func (s *DeleteOpContext) GetParser() antlr.Parser { return s.parser }

func (s *DeleteOpContext) SimpleDeleteOp() ISimpleDeleteOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISimpleDeleteOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISimpleDeleteOpContext)
}

func (s *DeleteOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DeleteOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DeleteOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterDeleteOp(s)
	}
}

func (s *DeleteOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitDeleteOp(s)
	}
}

func (p *WorkloadParser) DeleteOp() (localctx IDeleteOpContext) {
	this := p
	_ = this

	localctx = NewDeleteOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, WorkloadParserRULE_deleteOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(81)
		p.SimpleDeleteOp()
	}

	return localctx
}

// ISimpleDeleteOpContext is an interface to support dynamic dispatch.
type ISimpleDeleteOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSimpleDeleteOpContext differentiates from other interfaces.
	IsSimpleDeleteOpContext()
}

type SimpleDeleteOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySimpleDeleteOpContext() *SimpleDeleteOpContext {
	var p = new(SimpleDeleteOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_simpleDeleteOp
	return p
}

func (*SimpleDeleteOpContext) IsSimpleDeleteOpContext() {}

func NewSimpleDeleteOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SimpleDeleteOpContext {
	var p = new(SimpleDeleteOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_simpleDeleteOp

	return p
}

func (s *SimpleDeleteOpContext) GetParser() antlr.Parser { return s.parser }

func (s *SimpleDeleteOpContext) TableOrRowRef() ITableOrRowRefContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITableOrRowRefContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITableOrRowRefContext)
}

func (s *SimpleDeleteOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SimpleDeleteOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SimpleDeleteOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterSimpleDeleteOp(s)
	}
}

func (s *SimpleDeleteOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitSimpleDeleteOp(s)
	}
}

func (p *WorkloadParser) SimpleDeleteOp() (localctx ISimpleDeleteOpContext) {
	this := p
	_ = this

	localctx = NewSimpleDeleteOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, WorkloadParserRULE_simpleDeleteOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(83)
		p.Match(WorkloadParserT__7)
	}
	{
		p.SetState(84)
		p.TableOrRowRef()
	}

	return localctx
}

// IRandomDmlOpContext is an interface to support dynamic dispatch.
type IRandomDmlOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsRandomDmlOpContext differentiates from other interfaces.
	IsRandomDmlOpContext()
}

type RandomDmlOpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRandomDmlOpContext() *RandomDmlOpContext {
	var p = new(RandomDmlOpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_randomDmlOp
	return p
}

func (*RandomDmlOpContext) IsRandomDmlOpContext() {}

func NewRandomDmlOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RandomDmlOpContext {
	var p = new(RandomDmlOpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_randomDmlOp

	return p
}

func (s *RandomDmlOpContext) GetParser() antlr.Parser { return s.parser }

func (s *RandomDmlOpContext) TableName() antlr.TerminalNode {
	return s.GetToken(WorkloadParserTableName, 0)
}

func (s *RandomDmlOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RandomDmlOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RandomDmlOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterRandomDmlOp(s)
	}
}

func (s *RandomDmlOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitRandomDmlOp(s)
	}
}

func (p *WorkloadParser) RandomDmlOp() (localctx IRandomDmlOpContext) {
	this := p
	_ = this

	localctx = NewRandomDmlOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, WorkloadParserRULE_randomDmlOp)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(86)
		p.Match(WorkloadParserT__8)
	}
	{
		p.SetState(87)
		p.Match(WorkloadParserTableName)
	}

	return localctx
}

// ITableOrRowRefContext is an interface to support dynamic dispatch.
type ITableOrRowRefContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsTableOrRowRefContext differentiates from other interfaces.
	IsTableOrRowRefContext()
}

type TableOrRowRefContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTableOrRowRefContext() *TableOrRowRefContext {
	var p = new(TableOrRowRefContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = WorkloadParserRULE_tableOrRowRef
	return p
}

func (*TableOrRowRefContext) IsTableOrRowRefContext() {}

func NewTableOrRowRefContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TableOrRowRefContext {
	var p = new(TableOrRowRefContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = WorkloadParserRULE_tableOrRowRef

	return p
}

func (s *TableOrRowRefContext) GetParser() antlr.Parser { return s.parser }

func (s *TableOrRowRefContext) TableName() antlr.TerminalNode {
	return s.GetToken(WorkloadParserTableName, 0)
}

func (s *TableOrRowRefContext) RowID() antlr.TerminalNode {
	return s.GetToken(WorkloadParserRowID, 0)
}

func (s *TableOrRowRefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableOrRowRefContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TableOrRowRefContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.EnterTableOrRowRef(s)
	}
}

func (s *TableOrRowRefContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(WorkloadListener); ok {
		listenerT.ExitTableOrRowRef(s)
	}
}

func (p *WorkloadParser) TableOrRowRef() (localctx ITableOrRowRefContext) {
	this := p
	_ = this

	localctx = NewTableOrRowRefContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, WorkloadParserRULE_tableOrRowRef)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(89)
		_la = p.GetTokenStream().LA(1)

		if !(_la == WorkloadParserTableName || _la == WorkloadParserRowID) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}
