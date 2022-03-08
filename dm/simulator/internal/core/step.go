package core

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/pingcap/errors"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
	"go.uber.org/zap"
)

type DMLWorkloadStep interface {
	Execute(*DMLWorkloadStepContext) error
	GetTableName() string
}

type DMLWorkloadStepContext struct {
	tx      *sql.Tx
	ctx     context.Context
	mcp     *sqlgen.ModificationCandidatePool
	rowRefs map[string]*sqlgen.UniqueKey
}

type insertStep struct {
	assignedRowID string
	sqlGen        sqlgen.SQLGenerator
	tableName     string
}

func (stp *insertStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: INSERT, Table: %s, AssignedRowID: %s }",
		stp, stp.GetTableName(), stp.assignedRowID,
	)
}

func (stp *insertStep) GetTableName() string {
	return stp.tableName
}

func (stp *insertStep) Execute(sctx *DMLWorkloadStepContext) error {
	var err error
	sql, uk, err := stp.sqlGen.GenInsertRow()
	if err != nil {
		return errors.Annotate(err, "generate INSERT SQL error")
	}
	uk.OPLock.Lock()
	defer uk.OPLock.Unlock()
	_, err = sctx.tx.ExecContext(sctx.ctx, sql)
	if err != nil {
		return errors.Annotate(err, "execute INSERT SQL error")
	}
	err = sctx.mcp.AddUK(uk)
	if err != nil {
		return errors.Annotate(err, "add new UK to MCP error")
	}
	if len(stp.assignedRowID) > 0 {
		sctx.rowRefs[stp.assignedRowID] = uk

	}
	return nil
}

type updateStep struct {
	sqlGen          sqlgen.SQLGenerator
	assignmentRowID string
	inputRowID      string
	tableName       string
}

func (stp *updateStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: UPDATE, Table: %s, AssignedRowID: %s, InputRowID: %s }",
		stp, stp.GetTableName(), stp.assignmentRowID, stp.inputRowID,
	)
}

func (stp *updateStep) GetTableName() string {
	return stp.tableName
}

func (stp *updateStep) Execute(sctx *DMLWorkloadStepContext) error {
	var (
		err error
		uk  *sqlgen.UniqueKey
		ok  bool
	)
	if len(stp.inputRowID) > 0 {
		uk, ok = sctx.rowRefs[stp.inputRowID]
		if !ok {
			err = ErrRowIDNotFound
			plog.L().Error(err.Error(), zap.String("input_row_id", stp.inputRowID))
			return err
		}
	} else {
		uk = sctx.mcp.NextUK()
		if uk == nil {
			return ErrNoMCPData
		}
	}
	uk.OPLock.Lock()
	defer uk.OPLock.Unlock()
	sql, err := stp.sqlGen.GenUpdateRow(uk)
	if err != nil {
		return errors.Annotate(err, "generate UPDATE SQL error")
	}
	_, err = sctx.tx.ExecContext(sctx.ctx, sql)
	if err != nil {
		return errors.Annotate(err, "execute UPDATE SQL error")
	}
	if len(stp.assignmentRowID) > 0 {
		if stp.assignmentRowID != stp.inputRowID {
			sctx.rowRefs[stp.assignmentRowID] = uk
		}
	}
	return nil
}

type deleteStep struct {
	sqlGen     sqlgen.SQLGenerator
	inputRowID string
	tableName  string
}

func (stp *deleteStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: DELETE, Table: %s, InputRowID: %s }",
		stp, stp.GetTableName(), stp.inputRowID,
	)
}

func (stp *deleteStep) GetTableName() string {
	return stp.tableName
}

func (stp *deleteStep) Execute(sctx *DMLWorkloadStepContext) error {
	var (
		err error
		uk  *sqlgen.UniqueKey
		ok  bool
	)
	if len(stp.inputRowID) > 0 {
		uk, ok = sctx.rowRefs[stp.inputRowID]
		if !ok {
			err = ErrRowIDNotFound
			plog.L().Error(err.Error(), zap.String("input_row_id", stp.inputRowID))
			return err
		}
	} else {
		uk = sctx.mcp.NextUK()
		if uk == nil {
			return ErrNoMCPData
		}
	}
	uk.OPLock.Lock()
	defer uk.OPLock.Unlock()

	if len(stp.inputRowID) > 0 {
		delete(sctx.rowRefs, stp.inputRowID)
	}
	sql, err := stp.sqlGen.GenDeleteRow(uk)
	if err != nil {
		return errors.Annotate(err, "generate DELETE SQL error")
	}
	_, err = sctx.tx.ExecContext(sctx.ctx, sql)
	if err != nil {
		return errors.Annotate(err, "execute DELETE SQL error")
	}
	err = sctx.mcp.DeleteUK(uk)
	if err != nil {
		return errors.Annotate(err, "delete UK from MCP error")
	}
	return nil
}

type randomDMLStep struct {
	sqlGen    sqlgen.SQLGenerator
	tableName string
}

func (stp *randomDMLStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: RANDOM , Table: %s }",
		stp, stp.GetTableName(),
	)
}

func (stp *randomDMLStep) GetTableName() string {
	return stp.tableName
}

func (stp *randomDMLStep) Execute(sctx *DMLWorkloadStepContext) error {
	var err error
	dmlType := randType()
	switch dmlType {
	case sqlgen.INSERT_DMLType:
		theInsertStep := &insertStep{
			sqlGen: stp.sqlGen,
		}
		err = theInsertStep.Execute(sctx)
		if err != nil {
			return errors.Annotate(err, "simulate INSERT error")
		}
	case sqlgen.DELETE_DMLType:
		theDeleteStep := &deleteStep{
			sqlGen: stp.sqlGen,
		}
		err = theDeleteStep.Execute(sctx)
		if err != nil {
			return errors.Annotate(err, "simulate DELETE error")
		}
	default:
		theUpdateStep := &updateStep{
			sqlGen: stp.sqlGen,
		}
		err = theUpdateStep.Execute(sctx)
		if err != nil {
			return errors.Annotate(err, "simulate UPDATE error")
		}
	}
	return nil
}

func randType() sqlgen.DMLType {
	randNum := rand.Intn(4)
	switch randNum {
	case 0:
		return sqlgen.INSERT_DMLType
	case 1:
		return sqlgen.DELETE_DMLType
	default:
		return sqlgen.UPDATE_DMLType
	}
}
