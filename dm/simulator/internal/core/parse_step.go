package core

import (
	"strconv"

	"go.uber.org/zap"

	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/parser"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

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

func NewParseStepsListener(tblConfigs map[string]*config.TableConfig) *parseStepsListener {
	return &parseStepsListener{
		tblConfigs:        tblConfigs,
		rowIDTableNameMap: make(map[string]string),
	}
}

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

func (l *parseStepsListener) ExitRepeatStep(c *parser.RepeatStepContext) {
	parentRepeatCtx := l.curRepeatStepParsingCtx.prevRepeatedStepsCtx
	if parentRepeatCtx != nil {
		for i := 0; i < l.curRepeatStepParsingCtx.repeatTimes; i++ {
			for _, dmlStep := range l.curRepeatStepParsingCtx.totalSteps {
				parentRepeatCtx.totalSteps = append(
					parentRepeatCtx.totalSteps,
					dmlStep,
				)
			}
		}
	} else {
		for i := 0; i < l.curRepeatStepParsingCtx.repeatTimes; i++ {
			for _, dmlStep := range l.curRepeatStepParsingCtx.totalSteps {
				l.totalSteps = append(
					l.totalSteps,
					dmlStep,
				)
			}
		}
	}
	l.curRepeatStepParsingCtx = parentRepeatCtx
}

func (l *parseStepsListener) EnterDmlStep(c *parser.DmlStepContext) {
	l.curDMLParsingCtx = new(dmlParsingContext)
}

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

//assume parsing ctx is not nil
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
		return &insertStep{
			sqlGen:        sqlGen,
			assignedRowID: l.curDMLParsingCtx.AssignedRowID,
		}
	case "UPDATE":
		return &updateStep{
			sqlGen:          sqlGen,
			assignmentRowID: l.curDMLParsingCtx.AssignedRowID,
			inputRowID:      l.curDMLParsingCtx.InputRowID,
		}
	case "DELETE":
		return &deleteStep{
			sqlGen:     sqlGen,
			inputRowID: l.curDMLParsingCtx.InputRowID,
		}
	case "RANDOM":
		return &randomDMLStep{
			sqlGen: sqlGen,
		}
	default:
		return nil
	}
}

func (l *parseStepsListener) EnterInsertOp(c *parser.InsertOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "INSERT"
}

func (l *parseStepsListener) ExitAssignmentInsertOp(c *parser.AssignmentInsertOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.AssignedRowID = c.RowID().GetSymbol().GetText()
	if len(l.curDMLParsingCtx.TableName) > 0 {
		l.rowIDTableNameMap[l.curDMLParsingCtx.AssignedRowID] = l.curDMLParsingCtx.TableName
	}
}

func (l *parseStepsListener) EnterSimpleInsertOp(c *parser.SimpleInsertOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.TableName = c.TableName().GetSymbol().GetText()
}

func (l *parseStepsListener) EnterUpdateOp(c *parser.UpdateOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "UPDATE"
}

func (l *parseStepsListener) ExitAssignmentUpdateOp(c *parser.AssignmentUpdateOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.AssignedRowID = c.RowID().GetSymbol().GetText()
	if len(l.curDMLParsingCtx.TableName) > 0 {
		l.rowIDTableNameMap[l.curDMLParsingCtx.AssignedRowID] = l.curDMLParsingCtx.TableName
	}
}

func (l *parseStepsListener) EnterDeleteOp(c *parser.DeleteOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "DELETE"
}

func (l *parseStepsListener) EnterRandomDmlOp(c *parser.RandomDmlOpContext) {
	if l.curDMLParsingCtx == nil {
		return
	}
	l.curDMLParsingCtx.ExecType = "RANDOM"
	l.curDMLParsingCtx.TableName = c.TableName().GetSymbol().GetText()
}

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
