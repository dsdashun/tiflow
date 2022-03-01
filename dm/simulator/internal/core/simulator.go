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

	"go.uber.org/zap"

	"github.com/chaos-mesh/go-sqlsmith/types"
	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type simulatorImpl struct {
	sqlGen           sqlgen.SQLGenerator
	db               *sql.DB
	mcp              *sqlgen.ModificationCandidatePool
	totalExecutedTrx uint64
}

func NewSimulatorImpl(db *sql.DB, sqlGen sqlgen.SQLGenerator) *simulatorImpl {
	return &simulatorImpl{
		db:     db,
		sqlGen: sqlGen,
		mcp:    sqlgen.NewModificationCandidatePool(),
	}
}

func (s *simulatorImpl) DoSimulation(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.L().Info("context expired, simulation terminated")
			return
		default:
			// continue
		}
		s.simulateTrx(ctx)
	}
}

func (s *simulatorImpl) loadMCP(ctx context.Context) error {
	var err error
	sql, colMetas, err := s.sqlGen.GenLoadUniqueKeySQL()
	if err != nil {
		log.L().Error("generate load unique key SQL error", zap.Error(err))
		return err
	}
	rows, err := s.db.QueryContext(ctx, sql)
	if err != nil {
		log.L().Error("execute load UK SQL error", zap.Error(err))
		return err
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
				return ErrUnsupportedColumnType
			}
			values = append(values, valHolder)
		}
		err = rows.Scan(values...)
		if err != nil {
			log.L().Error("scan values error", zap.Error(err))
			return err
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
		log.L().Error("fetch rows has error", zap.Error(err))
		return err
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

func (s *simulatorImpl) prepareData(ctx context.Context, recordCount int) error {
	var (
		err error
		sql string
	)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	sql, err = s.sqlGen.GenTruncateTable()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return err
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

func (s *simulatorImpl) simulateTrx(ctx context.Context) error {
	var err error
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	dmlType := randType()
	switch dmlType {
	case sqlgen.INSERT_DMLType:
		_, err = s.simulateInsert(ctx, tx)
		if err != nil {
			log.L().Error("simulate INSERT error", zap.Error(err))
			tx.Rollback()
			return err
		}
	case sqlgen.DELETE_DMLType:
		err = s.simulateDelete(ctx, tx)
		if err != nil {
			log.L().Error("simulate DELETE error", zap.Error(err))
			tx.Rollback()
			return err
		}
	default:
		err = s.simulateUpdate(ctx, tx)
		if err != nil {
			log.L().Error("simulate UPDATE error", zap.Error(err))
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.L().Error("transaction COMMIT error", zap.Error(err))
		return err
	}
	atomic.AddUint64(&s.totalExecutedTrx, 1)
	return nil
}

func (s *simulatorImpl) simulateInsert(ctx context.Context, tx *sql.Tx) (*sqlgen.UniqueKey, error) {
	var err error
	sql, uk, err := s.sqlGen.GenInsertRow()
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	err = s.mcp.AddUK(uk)
	if err != nil {
		return nil, err
	}
	return uk, nil
}

func (s *simulatorImpl) simulateDelete(ctx context.Context, tx *sql.Tx) error {
	var err error
	uk := s.mcp.NextUK()
	if uk == nil {
		return ErrNoMCPData
	}
	sql, err := s.sqlGen.GenDeleteRow(uk)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return err
	}
	err = s.mcp.DeleteUK(uk)
	if err != nil {
		return err
	}
	return nil
}

func (s *simulatorImpl) simulateUpdate(ctx context.Context, tx *sql.Tx) error {
	var err error
	uk := s.mcp.NextUK()
	if uk == nil {
		return ErrNoMCPData
	}
	sql, err := s.sqlGen.GenUpdateRow(uk)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, sql)
	if err != nil {
		return err
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
