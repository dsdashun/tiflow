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
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pingcap/errors"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
	"github.com/pingcap/tiflow/dm/simulator/internal/schema"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
	"github.com/pingcap/tiflow/dm/simulator/internal/workload"
)

// DBSimulator is the core simulator execution framework.
// One DBSimulator concentrates on simulating a DB instance.
type DBSimulator struct {
	sync.RWMutex
	isRunning          atomic.Bool
	wg                 sync.WaitGroup
	workloadLock       sync.RWMutex
	workloadSimulators map[string]workload.WorkloadSimulator
	ctx                context.Context
	cancel             func()
	workerCh           chan workload.WorkloadSimulator
	db                 *sql.DB
	tblConfigs         map[string]*config.TableConfig
	mcpMap             map[string]*mcp.ModificationCandidatePool
}

// NewDBSimulator creates a new DB simulator.
func NewDBSimulator(db *sql.DB, tblConfigs map[string]*config.TableConfig) *DBSimulator {
	return &DBSimulator{
		db:                 db,
		tblConfigs:         tblConfigs,
		workloadSimulators: make(map[string]workload.WorkloadSimulator),
		mcpMap:             make(map[string]*mcp.ModificationCandidatePool),
	}
}

// workerFn is the main loop for a simulation worker.
// It will continuously receive a workload and simulate a transaction for it.
func (s *DBSimulator) workerFn(ctx context.Context, workloadCh <-chan workload.WorkloadSimulator) {
	for {
		select {
		case <-ctx.Done():
			plog.L().Info("worker context is terminated")
			return
		case theWorkload := <-workloadCh:
			err := theWorkload.SimulateTrx(ctx, s.db, s.mcpMap)
			if err != nil {
				plog.L().Error("simulate a trx error", zap.Error(err))
			}
		case <-time.After(100 * time.Millisecond):
			// continue on
		}
	}
}

// AddWorkload adds a workload simulator to the DB simulator.
func (s *DBSimulator) AddWorkload(workloadName string, ts workload.WorkloadSimulator) {
	s.workloadLock.Lock()
	defer s.workloadLock.Unlock()
	s.workloadSimulators[workloadName] = ts
}

// RemoveWorkload removes a workload from the DB simulator given a workload name.
func (s *DBSimulator) RemoveWorkload(workloadName string) {
	s.workloadLock.Lock()
	defer s.workloadLock.Unlock()
	delete(s.workloadSimulators, workloadName)
}

func (s *DBSimulator) getAllInvolvedTableConfigs() (map[string]*config.TableConfig, error) {
	allInvolvedTblConfigs := make(map[string]*config.TableConfig)
	for _, workloadSimu := range s.workloadSimulators {
		involvedTableNames := workloadSimu.GetInvolvedTables()
		for _, tblName := range involvedTableNames {
			if _, ok := allInvolvedTblConfigs[tblName]; ok {
				continue
			}
			cfg, ok := s.tblConfigs[tblName]
			if !ok {
				err := ErrTableConfigNotFound
				plog.L().Error(err.Error(), zap.String("table_id", tblName))
				return nil, err
			}
			allInvolvedTblConfigs[tblName] = cfg
		}
	}
	return allInvolvedTblConfigs, nil
}

// PrepareData prepares data for all the involving tables.
// It will fill all the involved tables with the specified number of records.
// Currently, all the existing data of the involved tables will be truncated before preparation.
// It should be called after all the workload simulators have been added,
// because the DB simulator will collect all the involved tables for preparation.
func (s *DBSimulator) PrepareData(ctx context.Context, recordCount int) error {
	allInvolvedTblConfigs, err := s.getAllInvolvedTableConfigs()
	if err != nil {
		return errors.Annotate(err, "prepare data error")
	}
	for _, tblConf := range allInvolvedTblConfigs {
		var (
			err error
			sql string
		)
		sqlGen := sqlgen.NewSQLGeneratorImpl(tblConf)
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return errors.Annotate(err, "begin trx for preparing data error")
		}
		sql, err = sqlGen.GenTruncateTable()
		if err != nil {
			return errors.Annotate(err, "generate truncate table SQL error")
		}
		_, err = tx.ExecContext(ctx, sql)
		if err != nil {
			return errors.Annotate(err, "execute truncate table SQL error")
		}
		for i := 0; i < recordCount; i++ {
			sql, _, err = sqlGen.GenInsertRow()
			if err != nil {
				plog.L().Error("generate INSERT SQL error", zap.Error(err))
				continue
			}
			_, err = tx.ExecContext(ctx, sql)
			if err != nil {
				plog.L().Error("execute SQL error", zap.Error(err))
				continue
			}
		}
		err = tx.Commit()
		if err != nil {
			return errors.Annotate(err, "commit trx error")
		}
	}
	return nil
}

// LoadMCP loads the unique keys of all the involving tables into modification candidate pools (MCPs).
// It should be called after the data have been prepared.
func (s *DBSimulator) LoadMCP(ctx context.Context) error {
	allInvolvedTblConfigs, err := s.getAllInvolvedTableConfigs()
	if err != nil {
		return errors.Annotate(err, "load MCP error")
	}
	for tblName, tblConf := range allInvolvedTblConfigs {
		theMCP, err := s.NewMCPFromTable(ctx, tblConf)
		if err != nil {
			return errors.Annotate(err, "new MCP from table error")
		}
		s.mcpMap[tblName] = theMCP
	}
	return nil
}

func (s *DBSimulator) NewMCPFromTable(ctx context.Context, tblConf *config.TableConfig) (*mcp.ModificationCandidatePool, error) {
	sqlGen := sqlgen.NewSQLGeneratorImpl(tblConf)
	sql, colMetas, err := sqlGen.GenLoadUniqueKeySQL()
	if err != nil {
		return nil, errors.Annotate(err, "generate load unique key SQL error")
	}
	rows, err := s.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, errors.Annotate(err, "execute load Unique SQL error")
	}
	defer rows.Close()
	theMCP := mcp.NewModificationCandidatePool(8192)
	for rows.Next() {
		values := make([]interface{}, 0)
		for _, colMeta := range colMetas {
			valHolder := newColValueHolder(colMeta)
			if valHolder == nil {
				plog.L().Error("unsupported data type",
					zap.String("column_name", colMeta.ColumnName),
					zap.String("data_type", colMeta.DataType),
				)
				return nil, errors.Trace(ErrUnsupportedColumnType)
			}
			values = append(values, valHolder)
		}
		err = rows.Scan(values...)
		if err != nil {
			return nil, errors.Annotate(err, "scan values error")
		}
		ukValue := make(map[string]interface{})
		for i, v := range values {
			colMeta := colMetas[i]
			ukValue[colMeta.ColumnName] = getValueHolderValue(v)
		}
		theUK := mcp.NewUniqueKey(-1, ukValue)
		if addErr := theMCP.AddUK(theUK); addErr != nil {
			plog.L().Error("add UK into MCP error", zap.Error(addErr), zap.String("unique_key", theUK.String()))
		}
		plog.L().Debug("add UK value to the pool", zap.Any("uk", theUK))
	}
	if rows.Err() != nil {
		return nil, errors.Annotate(err, "fetch rows has error")
	}
	return theMCP, nil
}

func newColValueHolder(colMeta *config.ColumnDefinition) interface{} {
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

// StartSimulation starts simulation of this DB simulator.
// It will spawn several worker goroutines for handling workloads in parallel.
func (s *DBSimulator) StartSimulation(ctx context.Context) error {
	if s.isRunning.Load() {
		plog.L().Info("the DB simulator has already been started")
		return nil
	}
	return func() error {
		s.Lock()
		defer s.Unlock()
		s.ctx, s.cancel = context.WithCancel(ctx)
		workerCount := 8 // currently, it is hard-coded.  TODO: make it a input parameter.
		s.wg.Add(workerCount)
		s.workerCh = make(chan workload.WorkloadSimulator, workerCount)
		for i := 0; i < workerCount; i++ {
			go func() {
				defer s.wg.Done()
				s.workerFn(s.ctx, s.workerCh)
				plog.L().Info("worker exit")
			}()
		}
		s.wg.Add(1)
		go func(ctx context.Context) {
			defer s.wg.Done()
			for s.isRunning.Load() {
				select {
				case <-ctx.Done():
					plog.L().Info("context is done")
					return
				default:
					s.DoSimulation(ctx)
				}
			}
		}(s.ctx)
		s.isRunning.Store(true)
		plog.L().Info("the DB simulator has been started")
		return nil
	}()
}

// StopSimulation stops simulation of this DB simulator.
func (s *DBSimulator) StopSimulation() error {
	if !s.isRunning.Load() {
		plog.L().Info("the server has already been closed")
		return nil
	}
	// atomic operations on closing the server
	plog.L().Info("begin to stop the DB simulator")
	func() {
		s.Lock()
		defer s.Unlock()
		s.cancel()
		plog.L().Info("begin to wait all the goroutines to finish")
		s.wg.Wait() // wait all sub-goroutines finished
		plog.L().Info("all the goroutines finished")
		s.isRunning.Store(false)
		plog.L().Info("the DB simulator is stopped")
	}()
	return nil
}

// DoSimulation is the main loop for this DB simulator.
// It will randomly choose a workload simulator and assign to a execution worker.
func (s *DBSimulator) DoSimulation(ctx context.Context) {
	var theWorkload workload.WorkloadSimulator
	for {
		select {
		case <-ctx.Done():
			plog.L().Info("context expired, simulation terminated")
			return
		default:
			theWorkload = func() workload.WorkloadSimulator {
				s.workloadLock.RLock()
				defer s.workloadLock.RUnlock()
				weightMap := make(map[string]int)
				for workloadName := range s.workloadSimulators {
					weightMap[workloadName] = 1
				}
				workloadName := randomChooseKeyByWeights(weightMap)
				return s.workloadSimulators[workloadName]
			}()
		}
		select {
		case s.workerCh <- theWorkload:
			// continue on
		case <-time.After(1 * time.Second):
			// timeout and continue on.  This is to prevent the blocking of sending a workload to the channel
			// When the program exits, that might happen
		}
	}
}

func randomChooseKeyByWeights(weights map[string]int) string {
	expandedKeys := []string{}
	for k, weight := range weights {
		for i := 0; i < weight; i++ {
			expandedKeys = append(expandedKeys, k)
		}
	}
	idx := rand.Intn(len(expandedKeys))
	return expandedKeys[idx]
}

func (s *DBSimulator) NewTableConfigFromDB(ctx context.Context, tableID string, databaseName string, tableName string) (*config.TableConfig, error) {
	theLogger := plog.L().With(
		zap.String("table_id", tableID),
		zap.String("db_name", databaseName),
		zap.String("table_name", tableName),
	)
	sg := schema.NewMySQLSchemaGetter(s.db, databaseName, tableName)
	newColDefs, err := sg.GetColumnDefinitions(ctx)
	if err != nil {
		errMsg := "get column definitions error"
		theLogger.Error(errMsg,
			zap.Error(err),
		)
		return nil, errors.Annotate(err, errMsg)
	}
	newUKColumns, err := sg.GetUniqueKeyColumns(ctx)
	if err != nil {
		errMsg := "get unique key error"
		theLogger.Error(errMsg,
			zap.Error(err),
		)
		return nil, errors.Annotate(err, errMsg)
	}
	return &config.TableConfig{
		TableID:              tableID,
		DatabaseName:         databaseName,
		TableName:            tableName,
		Columns:              newColDefs,
		UniqueKeyColumnNames: newUKColumns,
	}, nil
}

func (s *DBSimulator) RefreshTableSchema(ctx context.Context, tableID string) error {
	currentTblConfig, ok := s.tblConfigs[tableID]
	if !ok {
		errMsg := "table ID not found"
		plog.L().Error(errMsg, zap.String("table_id", tableID))
		return errors.New(errMsg)
	}
	newTableConfig, err := s.NewTableConfigFromDB(ctx, tableID, currentTblConfig.DatabaseName, currentTblConfig.TableName)
	if err != nil {
		errMsg := "new table config from DB error"
		plog.L().Error(errMsg,
			zap.Error(err),
			zap.String("table_id", tableID),
			zap.String("database_name", currentTblConfig.DatabaseName),
			zap.String("table_name", currentTblConfig.TableName),
		)
		return errors.Annotate(err, errMsg)
	}
	// TODO: the modification should has least impact on current simulation
	s.tblConfigs[tableID] = newTableConfig
	for _, ws := range s.workloadSimulators {
		ws.SetTableConfig(tableID, newTableConfig)
	}
	newMCP, err := s.NewMCPFromTable(ctx, newTableConfig)
	if err != nil {
		return errors.Annotate(err, "new MCP from table error")
	}
	s.mcpMap[tableID] = newMCP
	return nil
}
