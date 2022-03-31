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
	workloadSimulators map[string]workload.Simulator
	ctx                context.Context
	cancel             func()
	workerCh           chan workload.Simulator
	db                 *sql.DB
	tblConfigs         map[string]*config.TableConfig
	mcpMap             map[string]*mcp.ModificationCandidatePool
	sg                 schema.Getter
	mcpLoader          MCPLoader
}

// NewDBSimulator creates a new DB simulator.
func NewDBSimulator(db *sql.DB, tblConfigs map[string]*config.TableConfig) *DBSimulator {
	return &DBSimulator{
		db:                 db,
		tblConfigs:         tblConfigs,
		workloadSimulators: make(map[string]workload.Simulator),
		mcpMap:             make(map[string]*mcp.ModificationCandidatePool),
		sg:                 schema.NewMySQLSchemaGetter(db),
		mcpLoader:          NewMCPLoaderImpl(db),
	}
}

// GetTableConfig gets the table config of a table.
// It implements the `DBSimulatorInterface` interface.
func (s *DBSimulator) GetTableConfig(tableName string) *config.TableConfig {
	return s.tblConfigs[tableName]
}

// GetDB gets the *sql.DB object for the simulator.
// It implements the `DBSimulatorInterface` interface.
func (s *DBSimulator) GetDB() *sql.DB {
	return s.db
}

// GetContext gets the context of this simulator.
// It implements the `DBSimulatorInterface` interface.
func (s *DBSimulator) GetContext() context.Context {
	return s.ctx
}

// workerFn is the main loop for a simulation worker.
// It will continuously receive a workload and simulate a transaction for it.
func (s *DBSimulator) workerFn(ctx context.Context, workloadCh <-chan workload.Simulator) {
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
func (s *DBSimulator) AddWorkload(workloadName string, ts workload.Simulator) {
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
		return errors.Annotate(err, "get all involved table configs error")
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
		theMCP, err := s.mcpLoader.LoadMCP(ctx, tblConf)
		if err != nil {
			return errors.Annotate(err, "new MCP from table error")
		}
		s.mcpMap[tblName] = theMCP
	}
	return nil
}

// StartSimulation starts simulation of this DB simulator.
// It will spawn several worker goroutines for handling workloads in parallel.
// It implements the `DBSimulatorInterface` interface.
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
		s.workerCh = make(chan workload.Simulator, workerCount)
		for i := 0; i < workerCount; i++ {
			go func() {
				defer s.wg.Done()
				s.workerFn(s.ctx, s.workerCh)
				plog.L().Info("worker exit")
			}()
		}
		s.isRunning.Store(true) // set to running, so that the following goroutines can start the loop
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
		s.wg.Add(1)
		go func(ctx context.Context) {
			defer s.wg.Done()
			for s.isRunning.Load() {
				select {
				case <-ctx.Done():
					plog.L().Info("context is done")
					return
				default:
					plog.L().Info("begin to refresh all table schemas")
					if err := s.RefreshAllTableSchemas(ctx); err != nil {
						plog.L().Error("refresh all tables schemas error", zap.Error(err))
					}
					plog.L().Info("refresh all table schemas finished")
				}
				time.Sleep(3 * time.Second)
			}
		}(s.ctx)
		plog.L().Info("the DB simulator has been started")
		return nil
	}()
}

// StopSimulation stops simulation of this DB simulator.
// It implements the `DBSimulatorInterface` interface.
func (s *DBSimulator) StopSimulation() error {
	if !s.isRunning.Load() {
		plog.L().Info("the server has already been closed")
		return nil
	}
	// atomic operations on closing the server
	plog.L().Info("begin to stop the DB simulator")
	s.Lock()
	defer s.Unlock()
	s.cancel()
	plog.L().Info("begin to wait all the goroutines to finish")
	s.wg.Wait() // wait all sub-goroutines finished
	plog.L().Info("all the goroutines finished")
	s.isRunning.Store(false)
	plog.L().Info("the DB simulator is stopped")
	return nil
}

// DoSimulation is the main loop for this DB simulator.
// It will randomly choose a workload simulator and assign to a execution worker.
func (s *DBSimulator) DoSimulation(ctx context.Context) {
	var theWorkload workload.Simulator
	for {
		select {
		case <-ctx.Done():
			plog.L().Info("context expired, simulation terminated")
			return
		default:
			theWorkload = func() workload.Simulator {
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
		if !theWorkload.IsEnabled() {
			continue
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

// NewTableConfigFromDB generates a new table config from the database.
func (s *DBSimulator) NewTableConfigFromDB(ctx context.Context, tableID string, databaseName string, tableName string) (*config.TableConfig, error) {
	theLogger := plog.L().With(
		zap.String("table_id", tableID),
		zap.String("db_name", databaseName),
		zap.String("table_name", tableName),
	)
	newColDefs, err := s.sg.GetColumnDefinitions(ctx, databaseName, tableName)
	if err != nil {
		errMsg := "get column definitions error"
		theLogger.Error(errMsg,
			zap.Error(err),
		)
		return nil, errors.Annotate(err, errMsg)
	}
	newUKColumns, err := s.sg.GetUniqueKeyColumns(ctx, databaseName, tableName)
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

// RefreshTableSchema refreshes the table schema of a table in the DB simulator.
// It will not only updates the table config, but also refresh the MCP.
func (s *DBSimulator) RefreshTableSchema(ctx context.Context, tableID string) error {
	currentTblConfig, ok := s.tblConfigs[tableID]
	if !ok {
		errMsg := "table ID not found"
		plog.L().Error(errMsg, zap.String("table_id", tableID))
		return errors.New(errMsg)
	}
	theLogger := plog.L().With(
		zap.String("table_id", tableID),
		zap.String("database_name", currentTblConfig.DatabaseName),
		zap.String("table_name", currentTblConfig.TableName),
	)
	newTableConfig, err := s.NewTableConfigFromDB(ctx, tableID, currentTblConfig.DatabaseName, currentTblConfig.TableName)
	if err != nil {
		errMsg := "new table config from DB error"
		theLogger.Error(errMsg,
			zap.Error(err),
		)
		return errors.Annotate(err, errMsg)
	}
	// If newTableConfig deep equal to the current one, do nothing.
	if currentTblConfig.IsDeepEqual(newTableConfig) {
		return nil
	}
	theLogger.Info("the table schema has changed")
	// disable all the involved workloads of this table first
	for _, ws := range s.workloadSimulators {
		if ws.DoesInvolveTable(tableID) && ws.IsEnabled() {
			ws.Disable()
			defer ws.Enable()
		}
	}
	s.tblConfigs[tableID] = newTableConfig
	for _, ws := range s.workloadSimulators {
		if setErr := ws.SetTableConfig(tableID, newTableConfig); setErr != nil {
			return errors.Annotate(setErr, "set table config error")
		}
	}
	newMCP, err := s.mcpLoader.LoadMCP(ctx, newTableConfig)
	if err != nil {
		return errors.Annotate(err, "new MCP from table error")
	}
	s.mcpMap[tableID] = newMCP
	return nil
}

// RefreshTableSchema refreshes all the table schemas in the DB simulator.
func (s *DBSimulator) RefreshAllTableSchemas(ctx context.Context) error {
	allInvolvedTblConfigs, err := s.getAllInvolvedTableConfigs()
	if err != nil {
		return errors.Annotate(err, "get all involved table configs error")
	}
	for tableID := range allInvolvedTblConfigs {
		if err := s.RefreshTableSchema(ctx, tableID); err != nil {
			errMsg := "refresh table schema error"
			plog.L().Error(errMsg, zap.Error(err), zap.String("table_id", tableID))
			return errors.Annotate(err, errMsg)
		}
	}
	return nil
}

// RefreshTableSchema loads all the table schemas in the DB simulator.
// It will only updates the table config, not refreshing the MCP.
// It is used at the start of the DB simulator.
func (s *DBSimulator) LoadAllTableSchemas(ctx context.Context) error {
	allInvolvedTblConfigs, err := s.getAllInvolvedTableConfigs()
	if err != nil {
		return errors.Annotate(err, "get all involved table configs error")
	}
	for tableID, theTblConfig := range allInvolvedTblConfigs {
		theLogger := plog.L().With(
			zap.String("table_id", tableID),
			zap.String("database_name", theTblConfig.DatabaseName),
			zap.String("table_name", theTblConfig.TableName),
		)
		newTableConfig, err := s.NewTableConfigFromDB(ctx, tableID, theTblConfig.DatabaseName, theTblConfig.TableName)
		if err != nil {
			errMsg := "new table config from DB error"
			theLogger.Error(errMsg,
				zap.Error(err),
			)
			return errors.Annotate(err, errMsg)
		}
		s.tblConfigs[tableID] = newTableConfig
		for _, ws := range s.workloadSimulators {
			if err := ws.SetTableConfig(tableID, newTableConfig); err != nil {
				return errors.Annotate(err, "set table config error")
			}
		}
	}
	return nil
}

// Prepare does all the preparation works before the simulator starts.
// It implements the `DBSimulatorInterface` interface.
func (s *DBSimulator) Prepare(ctx context.Context) error {
	plog.L().Info("begin to load table schemas")
	if err := s.LoadAllTableSchemas(ctx); err != nil {
		return errors.Annotate(err, "load all table schemas error")
	}
	plog.L().Info("loading table schemas [DONE]")
	plog.L().Info("begin to prepare table data")
	if err := s.PrepareData(ctx, 4096); err != nil {
		return errors.Annotate(err, "prepare data error")
	}
	plog.L().Info("preparing table data [DONE]")
	plog.L().Info("begin to load UKs into MCP")
	if err := s.LoadMCP(ctx); err != nil {
		return errors.Annotate(err, "load data into MCP error")
	}
	plog.L().Info("loading UKs into MCP [DONE]")
	return nil
}
