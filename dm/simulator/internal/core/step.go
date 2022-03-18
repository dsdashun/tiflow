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
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/pingcap/errors"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
	"go.uber.org/zap"
)

// DMLWorkloadStep contains all the operations for a DML workload step.
type DMLWorkloadStep interface {
	// Execute executes the step in a workload.
	Execute(*DMLWorkloadStepContext) error

	// GetTableName returns the related table of this step.
	GetTableName() string
}

// DMLWorkloadStepContext is the context when a workload step is executed.
type DMLWorkloadStepContext struct {
	tx       *sql.Tx
	ctx      context.Context
	mcp      *mcp.ModificationCandidatePool
	rowRefs  map[string]*mcp.UniqueKey
	addedUKs map[string]map[*mcp.UniqueKey]struct{}
}

// InsertStep implements the workload step for INSERT.
type InsertStep struct {
	assignedRowID string
	sqlGen        sqlgen.SQLGenerator
	tableName     string
}

// String returns the string representation of an InsertStep.
func (stp *InsertStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: INSERT, Table: %s, AssignedRowID: %s }",
		stp, stp.GetTableName(), stp.assignedRowID,
	)
}

// GetTableName implements the DMLWorkloadStep interface for InsertStep.
func (stp *InsertStep) GetTableName() string {
	return stp.tableName
}

// Execute implements the DMLWorkloadStep interface for InsertStep.
// It is the core execution logic.
func (stp *InsertStep) Execute(sctx *DMLWorkloadStepContext) error {
	var err error
	sql, uk, err := stp.sqlGen.GenInsertRow()
	if err != nil {
		errMsg := "generate INSERT SQL error"
		plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", stp.GetTableName()))
		return errors.Annotate(err, errMsg)
	}
	_, err = sctx.tx.ExecContext(sctx.ctx, sql)
	if err != nil {
		errMsg := "execute INSERT SQL error"
		plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", stp.GetTableName()), zap.String("sql", sql))
		return errors.Annotate(err, errMsg)
	}
	if _, ok := sctx.addedUKs[stp.tableName]; !ok {
		sctx.addedUKs[stp.tableName] = make(map[*mcp.UniqueKey]struct{})
	}
	sctx.addedUKs[stp.tableName][uk] = struct{}{}
	if len(stp.assignedRowID) > 0 {
		sctx.rowRefs[stp.assignedRowID] = uk
	}
	return nil
}

// UpdateStep implements the workload step for UPDATE.
type UpdateStep struct {
	sqlGen          sqlgen.SQLGenerator
	assignmentRowID string
	inputRowID      string
	tableName       string
}

// String returns the string representation of an UpdateStep.
func (stp *UpdateStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: UPDATE, Table: %s, AssignedRowID: %s, InputRowID: %s }",
		stp, stp.GetTableName(), stp.assignmentRowID, stp.inputRowID,
	)
}

// GetTableName implements the DMLWorkloadStep interface for UpdateStep.
func (stp *UpdateStep) GetTableName() string {
	return stp.tableName
}

// Execute implements the DMLWorkloadStep interface for UpdateStep.
// It is the core execution logic.
func (stp *UpdateStep) Execute(sctx *DMLWorkloadStepContext) error {
	var (
		err error
		uk  *mcp.UniqueKey
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

	sql, err := stp.sqlGen.GenUpdateRow(uk)
	if err != nil {
		errMsg := "generate UPDATE SQL error"
		plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", stp.GetTableName()), zap.String("unique_key", uk.String()))
		return errors.Annotate(err, errMsg)
	}
	_, err = sctx.tx.ExecContext(sctx.ctx, sql)
	if err != nil {
		errMsg := "execute UPDATE SQL error"
		plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", stp.GetTableName()), zap.String("sql", sql))
		return errors.Annotate(err, errMsg)
	}
	if len(stp.assignmentRowID) > 0 {
		if stp.assignmentRowID != stp.inputRowID {
			sctx.rowRefs[stp.assignmentRowID] = uk
		}
	}
	return nil
}

// DeleteStep implements the workload step for DELETE.
type DeleteStep struct {
	sqlGen     sqlgen.SQLGenerator
	inputRowID string
	tableName  string
}

// String returns the string representation of a DeleteStep.
func (stp *DeleteStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: DELETE, Table: %s, InputRowID: %s }",
		stp, stp.GetTableName(), stp.inputRowID,
	)
}

// GetTableName implements the DMLWorkloadStep interface for DeleteStep.
func (stp *DeleteStep) GetTableName() string {
	return stp.tableName
}

// Execute implements the DMLWorkloadStep interface for DeleteStep.
// It is the core execution logic.
func (stp *DeleteStep) Execute(sctx *DMLWorkloadStepContext) error {
	var (
		err error
		uk  *mcp.UniqueKey
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

	if len(stp.inputRowID) > 0 {
		delete(sctx.rowRefs, stp.inputRowID)
	}
	hasDeletedFromTempUKs := false
	if _, ok := sctx.addedUKs[stp.tableName]; ok {
		if _, ok := sctx.addedUKs[stp.tableName][uk]; ok {
			delete(sctx.addedUKs[stp.tableName], uk)
			hasDeletedFromTempUKs = true
		}
	}
	if !hasDeletedFromTempUKs {
		err = sctx.mcp.DeleteUK(uk)
		if err != nil {
			errMsg := "delete UK from MCP error"
			plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", stp.GetTableName()), zap.String("unique_key", uk.String()))
			return errors.Annotate(err, errMsg)
		}
	}
	sql, err := stp.sqlGen.GenDeleteRow(uk)
	if err != nil {
		errMsg := "generate DELETE SQL error"
		plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", stp.GetTableName()), zap.String("unique_key", uk.String()))
		return errors.Annotate(err, errMsg)
	}
	_, err = sctx.tx.ExecContext(sctx.ctx, sql)
	if err != nil {
		errMsg := "execute DELETE SQL error"
		plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", stp.GetTableName()), zap.String("sql", sql))
		return errors.Annotate(err, errMsg)
	}
	return nil
}

// RandomDMLStep implements the workload step for RANDOM-DML.
type RandomDMLStep struct {
	sqlGen    sqlgen.SQLGenerator
	tableName string
}

// String returns the string representation of a RandomDMLStep.
func (stp *RandomDMLStep) String() string {
	return fmt.Sprintf(
		"%p: DML Step: { Type: RANDOM , Table: %s }",
		stp, stp.GetTableName(),
	)
}

// GetTableName implements the DMLWorkloadStep interface for RandomDMLStep.
func (stp *RandomDMLStep) GetTableName() string {
	return stp.tableName
}

// Execute implements the DMLWorkloadStep interface for RandomDMLStep.
// It is the core execution logic.
func (stp *RandomDMLStep) Execute(sctx *DMLWorkloadStepContext) error {
	var err error
	dmlType := randType()
	switch dmlType {
	case sqlgen.DMLTypeINSERT:
		theInsertStep := &InsertStep{
			sqlGen:    stp.sqlGen,
			tableName: stp.tableName,
		}
		err = theInsertStep.Execute(sctx)
		if err != nil {
			return errors.Annotate(err, "simulate INSERT error")
		}
	case sqlgen.DMLTypeDELETE:
		theDeleteStep := &DeleteStep{
			sqlGen:    stp.sqlGen,
			tableName: stp.tableName,
		}
		err = theDeleteStep.Execute(sctx)
		if err != nil {
			return errors.Annotate(err, "simulate DELETE error")
		}
	default:
		theUpdateStep := &UpdateStep{
			sqlGen:    stp.sqlGen,
			tableName: stp.tableName,
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
		return sqlgen.DMLTypeINSERT
	case 1:
		return sqlgen.DMLTypeDELETE
	default:
		return sqlgen.DMLTypeUPDATE
	}
}
