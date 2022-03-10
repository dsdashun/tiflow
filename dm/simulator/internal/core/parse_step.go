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

package core

import (
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/pingcap/errors"

	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/parser"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type parseStepsErrorListener struct {
	baseErrorListener antlr.ErrorListener
	parseErr          error
}

// NewParseStepsErrorListener generates a custom error listener for parsing workload steps.
// This listener will record the syntax error into an error object.
// It accepts an existing antlr.ErrorListener object as the base implementation.
func NewParseStepsErrorListener(el antlr.ErrorListener) *parseStepsErrorListener {
	return &parseStepsErrorListener{
		baseErrorListener: el,
	}
}

// Err returns the syntax error object from the error listener for parsing workload steps.
func (el *parseStepsErrorListener) Err() error {
	return el.parseErr
}

// SyntaxError implements the antlr.ErrorListener interface.
// It records the syntax error into an error object.
func (el *parseStepsErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	el.baseErrorListener.SyntaxError(recognizer, offendingSymbol, line, column, msg, e)
	el.parseErr = errors.Trace(
		fmt.Errorf("line %d:%d : %s", line, column, msg),
	)
}

// ReportAmbiguity implements the antlr.ErrorListener interface.
func (el *parseStepsErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	el.baseErrorListener.ReportAmbiguity(recognizer, dfa, startIndex, stopIndex, exact, ambigAlts, configs)
}

// ReportAttemptingFullContext implements the antlr.ErrorListener interface.
func (el *parseStepsErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	el.baseErrorListener.ReportAttemptingFullContext(recognizer, dfa, startIndex, stopIndex, conflictingAlts, configs)
}

// ReportContextSensitivity implements the antlr.ErrorListener interface.
func (el *parseStepsErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs antlr.ATNConfigSet) {
	el.baseErrorListener.ReportContextSensitivity(recognizer, dfa, startIndex, stopIndex, prediction, configs)
}

type repeatedStepsParsingContext struct {
	repeatTimes          int
	totalSteps           []DMLWorkloadStep
	prevRepeatedStepsCtx *repeatedStepsParsingContext
}

type dmlParsingContext struct {
	ExecType      string
	TableName     string
	AssignedRowID string
	InputRowID    string
}

type parseStepsListener struct {
	*parser.BaseWorkloadListener
	totalSteps              []DMLWorkloadStep
	curRepeatStepParsingCtx *repeatedStepsParsingContext
	curDMLParsingCtx        *dmlParsingContext
	tblConfigs              map[string]*config.TableConfig
	rowIDTableNameMap       map[string]string
}

// NewParseStepsListener generates a parser listener to trigger hooks on nodes of the parsed AST.
// The parse step listener will try to convert the AST into individual workload steps.
func NewParseStepsListener(tblConfigs map[string]*config.TableConfig) *parseStepsListener {
	return &parseStepsListener{
		tblConfigs:        tblConfigs,
		rowIDTableNameMap: make(map[string]string),
	}
}

// EnterRepeatStep is the hook when entering a 'repeatStep' node in AST.
// It will set a new 'repeat step context' for the listener.
func (l *parseStepsListener) EnterRepeatStep(c *parser.RepeatStepContext) {
	repeatTimes, err := strconv.Atoi(c.Int().GetSymbol().GetText())
	if err != nil {
		plog.L().Error("parse repeat step count into integer error", zap.Error(err))
		repeatTimes = 1
	}
	newRepeatStepCtx := &repeatedStepsParsingContext{
		repeatTimes:          repeatTimes,
		prevRepeatedStepsCtx: l.curRepeatStepParsingCtx,
	}
	l.curRepeatStepParsingCtx = newRepeatStepCtx
}

// ExitRepeatStep is the hook when leaving a 'repeatStep' node in AST.
// It will merge all the parsed workload steps into the final steps or its parent 'repeat step context',
// and then remove the current 'repeat step context'.
func (l *parseStepsListener) ExitRepeatStep(c *parser.RepeatStepContext) {
	parentRepeatCtx := l.curRepeatStepParsingCtx.prevRepeatedStepsCtx
	if parentRepeatCtx != nil {
		for i := 0; i < l.curRepeatStepParsingCtx.repeatTimes; i++ {
			parentRepeatCtx.totalSteps = append(
				parentRepeatCtx.totalSteps,
				l.curRepeatStepParsingCtx.totalSteps...,
			)
		}
	} else {
		for i := 0; i < l.curRepeatStepParsingCtx.repeatTimes; i++ {
			l.totalSteps = append(
				l.totalSteps,
				l.curRepeatStepParsingCtx.totalSteps...,
			)
		}
	}
	l.curRepeatStepParsingCtx = parentRepeatCtx
}

// EnterDmlStep is the hook when entering a 'dmlStep' node in AST.
// It will create a new 'DML step context' for the further processing.
func (l *parseStepsListener) EnterDmlStep(c *parser.DmlStepContext) {
	l.curDMLParsingCtx = new(dmlParsingContext)
}

// ExitDmlStep is the hook when leaving a 'dmlStep' node in AST.
// It will generate a DML workload step based on the execution type,
// and merge this workload step into the final steps or the current 'repeat step context'.
// At last, it will remove the current 'DML step context'.
func (l *parseStepsListener) ExitDmlStep(c *parser.DmlStepContext) {
	if l.curDMLParsingCtx == nil {
		return
	}

	dmlStep := l.newDMLWorkloadStep()
	if dmlStep == nil {
		plog.L().Warn("generated DML workload step is nil")
		return
	}
	if l.curRepeatStepParsingCtx != nil {
		l.curRepeatStepParsingCtx.totalSteps = append(
			l.curRepeatStepParsingCtx.totalSteps,
			dmlStep,
		)
	} else {
		l.totalSteps = append(
			l.totalSteps,
			dmlStep,
		)
	}
	l.curDMLParsingCtx = nil
}

// newDMLWorkloadStep generates a DML workload step based on different types.
func (l *parseStepsListener) newDMLWorkloadStep() DMLWorkloadStep {
	if l.curDMLParsingCtx == nil {
		return nil
	}
	tblConfig, ok := l.tblConfigs[l.curDMLParsingCtx.TableName]
	if !ok {
		plog.L().Error("cannot find the table config", zap.String("table_name", l.curDMLParsingCtx.TableName))
		return nil
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(tblConfig)

	switch l.curDMLParsingCtx.ExecType {
	case "INSERT":
		return &InsertStep{
			sqlGen:        sqlGen,
			tableName:     l.curDMLParsingCtx.TableName,
			assignedRowID: l.curDMLParsingCtx.AssignedRowID,
		}
	case "UPDATE":
		return &UpdateStep{
			sqlGen:          sqlGen,
			tableName:       l.curDMLParsingCtx.TableName,
			assignmentRowID: l.curDMLParsingCtx.AssignedRowID,
			inputRowID:      l.curDMLParsingCtx.InputRowID,
		}
	case "DELETE":
		return &DeleteStep{
			sqlGen:     sqlGen,
			tableName:  l.curDMLParsingCtx.TableName,
			inputRowID: l.curDMLParsingCtx.InputRowID,
		}
	case "RANDOM":
		return &RandomDMLStep{
			sqlGen:    sqlGen,
			tableName: l.curDMLParsingCtx.TableName,
		}
	default:
		return nil
	}
}

// EnterInsertOp is the hook when entering an 'insertOp' node in AST.
// It will set the execution type of the current 'DML step context' as 'INSERT'.
func (l *parseStepsListener) EnterInsertOp(c *parser.InsertOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "INSERT"
}

// ExitAssignmentInsertOp is the hook when leaving an 'assignmentInsertOp' node in AST.
// It will set the assignment relations of the current 'DML step context'.
func (l *parseStepsListener) ExitAssignmentInsertOp(c *parser.AssignmentInsertOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.AssignedRowID = c.RowID().GetSymbol().GetText()
	if len(l.curDMLParsingCtx.TableName) > 0 {
		l.rowIDTableNameMap[l.curDMLParsingCtx.AssignedRowID] = l.curDMLParsingCtx.TableName
	}
}

// EnterSimpleInsertOp is the hook when entering a 'simpleInsertOp' node in AST.
// It will set the affected table of the current 'DML step context'.
func (l *parseStepsListener) EnterSimpleInsertOp(c *parser.SimpleInsertOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.TableName = c.TableName().GetSymbol().GetText()
}

// EnterUpdateOp is the hook when entering an 'updateOp' node in AST.
// It will set the execution type of the current 'DML step context' as 'UPDATE'.
func (l *parseStepsListener) EnterUpdateOp(c *parser.UpdateOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "UPDATE"
}

// ExitAssignmentUpdateOp is the hook when leaving an 'assignmentUpdateOp' node in AST.
// It will set the assignment relations of the current 'DML step context'.
func (l *parseStepsListener) ExitAssignmentUpdateOp(c *parser.AssignmentUpdateOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.AssignedRowID = c.RowID().GetSymbol().GetText()
	if len(l.curDMLParsingCtx.TableName) > 0 {
		l.rowIDTableNameMap[l.curDMLParsingCtx.AssignedRowID] = l.curDMLParsingCtx.TableName
	}
}

// EnterDeleteOp is the hook when entering a 'deleteOp' node in AST.
// It will set the execution type of the current 'DML step context' as 'DELETE'.
func (l *parseStepsListener) EnterDeleteOp(c *parser.DeleteOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "DELETE"
}

// EnterRandomDmlOp is the hook when entering a 'randomDMLOp' node in AST.
// It will set the execution type of the current 'DML step context' as 'RANDOM',
// as well as setting the table name of the current DML parsing context.
func (l *parseStepsListener) EnterRandomDmlOp(c *parser.RandomDmlOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "RANDOM"
	l.curDMLParsingCtx.TableName = c.TableName().GetSymbol().GetText()
}

// EnterTableOrRowRef is the hook when entering a 'tableOrRowRef' node in AST.
// It will set the table name of the current DML parsing context,
// based on the input table name or the assignment relations for the row reference.
func (l *parseStepsListener) EnterTableOrRowRef(c *parser.TableOrRowRefContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	if c.TableName() != nil {
		l.curDMLParsingCtx.TableName = c.TableName().GetSymbol().GetText()
	} else {
		l.curDMLParsingCtx.InputRowID = c.RowID().GetSymbol().GetText()
		if tblName, ok := l.rowIDTableNameMap[l.curDMLParsingCtx.InputRowID]; ok {
			l.curDMLParsingCtx.TableName = tblName
		}
	}
}
