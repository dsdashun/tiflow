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
	"math/rand"
	"sync/atomic"

	"github.com/chaos-mesh/go-sqlsmith/types"
	"github.com/pingcap/errors"
	"go.uber.org/zap"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type workloadSimulatorImpl struct {
	sqlGen           sqlgen.SQLGenerator
	db               *sql.DB
	mcp              *sqlgen.ModificationCandidatePool
	totalExecutedTrx uint64
}

func NewWorkloadSimulatorImpl(db *sql.DB, sqlGen sqlgen.SQLGenerator) *workloadSimulatorImpl {
	return &workloadSimulatorImpl{
		db:     db,
		sqlGen: sqlGen,
		mcp:    sqlgen.NewModificationCandidatePool(),
	}
}

func (s *workloadSimulatorImpl) LoadMCP(ctx context.Context) error {
	var err error
	sql, colMetas, err := s.sqlGen.GenLoadUniqueKeySQL()
	if err != nil {
		return errors.Annotate(err, "generate load unique key SQL error")
	}
	rows, err := s.db.QueryContext(ctx, sql)
	if err != nil {
		return errors.Annotate(err, "execute load Unique SQL error")
	}
	s.mcp.Reset()
	for rows.Next() {
		values := make([]interface{}, 0)
		for _, colMeta := range colMetas {
			valHolder := newColValueHolder(colMeta)
			if valHolder == nil {
				log.L().Error("unsupported data type",
					zap.String("column_name", colMeta.Column),
					zap.String("data_type", colMeta.DataType),
				)
				return errors.Trace(ErrUnsupportedColumnType)
			}
			values = append(values, valHolder)
		}
		err = rows.Scan(values...)
		if err != nil {
			return errors.Annotate(err, "scan values error")
		}
		ukValue := make(map[string]interface{})
		for i, v := range values {
			colMeta := colMetas[i]
			ukValue[colMeta.Column] = getValueHolderValue(v)
		}
		theUK := &sqlgen.UniqueKey{
			RowID: -1,
			Value: ukValue,
		}
		s.mcp.AddUK(theUK)
		log.L().Debug("add UK value to the pool", zap.Any("uk", theUK))
	}
	if rows.Err() != nil {
		return errors.Annotate(err, "fetch rows has error")
	}
	return nil
}

func newColValueHolder(colMeta *types.Column) interface{} {
	switch colMeta.DataType {
	case "int":
		return new(int)
	case "varchar":
		return new(string)
	default:
		return nil
	}
}

func getValueHolderValue(valueHolder interface{}) interface{} {
	switch vh := valueHolder.(type) {
	case *int:
		return *vh
	case *string:
		return *vh
	default:
		return nil
	}
}

func (s *workloadSimulatorImpl) PrepareData(ctx context.Context, recordCount int) error {
	var (
		err error
		sql string
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Annotate(err, "begin trx for preparing data error")
	}
	sql, err = s.sqlGen.GenTruncateTable()
	if err != nil {
		return errors.Annotate(err, "generate truncate table SQL error")
	}
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return errors.Annotate(err, "execute truncate table SQL error")
	}
	for i := 0; i < recordCount; i++ {
		sql, _, err = s.sqlGen.GenInsertRow()
		if err != nil {
			log.L().Error("generate INSERT SQL error", zap.Error(err))
			continue
		}
		_, err = tx.ExecContext(ctx, sql)
		if err != nil {
			log.L().Error("execute SQL error", zap.Error(err))
			continue
		}
	}
	return tx.Commit()
}

func (s *workloadSimulatorImpl) SimulateTrx(ctx context.Context) error {
	var err error
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Annotate(err, "begin trx error when simulating a trx")
	}
	dmlType := randType()
	switch dmlType {
	case sqlgen.INSERT_DMLType:
		_, err = s.simulateInsert(ctx, tx)
		if err != nil {
			tx.Rollback()
			return errors.Annotate(err, "simulate INSERT error")
		}
	case sqlgen.DELETE_DMLType:
		err = s.simulateDelete(ctx, tx)
		if err != nil {
			tx.Rollback()
			return errors.Annotate(err, "simulate DELETE error")
		}
	default:
		err = s.simulateUpdate(ctx, tx)
		if err != nil {
			tx.Rollback()
			return errors.Annotate(err, "simulate UPDATE error")
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.Annotate(err, "trx COMMIT error when simulating a trx")
	}
	atomic.AddUint64(&s.totalExecutedTrx, 1)
	return nil
}

func (s *workloadSimulatorImpl) simulateInsert(ctx context.Context, tx *sql.Tx) (*sqlgen.UniqueKey, error) {
	var err error
	sql, uk, err := s.sqlGen.GenInsertRow()
	if err != nil {
		return nil, errors.Annotate(err, "generate INSERT SQL error")
	}
	uk.OPLock.Lock()
	defer uk.OPLock.Unlock()
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return nil, errors.Annotate(err, "execute INSERT SQL error")
	}
	err = s.mcp.AddUK(uk)
	if err != nil {
		return nil, errors.Annotate(err, "add new UK to MCP error")
	}
	return uk, nil
}

func (s *workloadSimulatorImpl) simulateDelete(ctx context.Context, tx *sql.Tx) error {
	var err error
	uk := s.mcp.NextUK()
	if uk == nil {
		return errors.Trace(ErrNoMCPData)
	}
	uk.OPLock.Lock()
	defer uk.OPLock.Unlock()
	sql, err := s.sqlGen.GenDeleteRow(uk)
	if err != nil {
		return errors.Annotate(err, "generate DELETE SQL error")
	}
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return errors.Annotate(err, "execute DELETE SQL error")
	}
	err = s.mcp.DeleteUK(uk)
	if err != nil {
		return errors.Annotate(err, "delete UK from MCP error")
	}
	return nil
}

func (s *workloadSimulatorImpl) simulateUpdate(ctx context.Context, tx *sql.Tx) error {
	var err error
	uk := s.mcp.NextUK()
	if uk == nil {
		return errors.Trace(ErrNoMCPData)
	}
	uk.OPLock.Lock()
	defer uk.OPLock.Unlock()
	sql, err := s.sqlGen.GenUpdateRow(uk)
	if err != nil {
		return errors.Annotate(err, "generate UPDATE SQL error")
	}
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return errors.Annotate(err, "execute UPDATE SQL error")
	}
	return nil
}

func randType() sqlgen.DMLType {
	randNum := rand.Int() % 4
	switch randNum {
	case 0:
		return sqlgen.INSERT_DMLType
	case 1:
		return sqlgen.DELETE_DMLType
	default:
		return sqlgen.UPDATE_DMLType
	}
}
