package core

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiflow/dm/pkg/log"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
	"github.com/pingcap/tiflow/dm/simulator/internal/utils"
)

type DBSimulator struct {
	sync.RWMutex
	isRunning          atomic.Bool
	wg                 sync.WaitGroup
	workloadLock       sync.RWMutex
	workloadSimulators map[string]WorkloadSimulator
	ctx                context.Context
	cancel             func()
	workerCh           chan WorkloadSimulator
	db                 *sql.DB
	tblConfigs         map[string]*config.TableConfig
	mcpMap             map[string]*sqlgen.ModificationCandidatePool
	opt                *dbSimulatorOption
}

type dbSimulatorOption struct {
	PrepareRecordCount int
}

func NewDefaultDBSimulatorOption() *dbSimulatorOption {
	return &dbSimulatorOption{
		PrepareRecordCount: 4096,
	}
}

type DBSimulatorOption interface {
	Apply(o *dbSimulatorOption)
}

type withPrepareRecordCount struct {
	recCount int
}

func (w *withPrepareRecordCount) Apply(o *dbSimulatorOption) {
	o.PrepareRecordCount = w.recCount
}

func WithPrepareRecordCount(recCount int) DBSimulatorOption {
	return &withPrepareRecordCount{
		recCount: recCount,
	}
}

func NewDBSimulator(db *sql.DB, tblConfigs map[string]*config.TableConfig, opts ...DBSimulatorOption) *DBSimulator {
	newOpt := NewDefaultDBSimulatorOption()
	for _, opt := range opts {
		opt.Apply(newOpt)
	}
	return &DBSimulator{
		db:                 db,
		tblConfigs:         tblConfigs,
		workloadSimulators: make(map[string]WorkloadSimulator),
		mcpMap:             make(map[string]*sqlgen.ModificationCandidatePool),
		opt:                newOpt,
	}
}

func (s *DBSimulator) workerFn(ctx context.Context, workloadCh <-chan WorkloadSimulator) {
	for {
		select {
		case <-ctx.Done():
			log.L().Info("worker context is terminated")
			return
		case theWorkload := <-workloadCh:
			err := theWorkload.SimulateTrx(ctx, s.db, s.mcpMap)
			if err != nil {
				log.L().Error("simulate a trx error", zap.Error(err))
			}
		case <-time.After(100 * time.Millisecond):
			//continue on
		}
	}
}

func (s *DBSimulator) AddWorkload(workloadName string, ts WorkloadSimulator) {
	s.workloadLock.Lock()
	defer s.workloadLock.Unlock()
	s.workloadSimulators[workloadName] = ts
}

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
			if cfg, ok := s.tblConfigs[tblName]; !ok {
				err := ErrTableConfigNotFound
				plog.L().Error(err.Error(), zap.String("table_id", tblName))
				return nil, err
			} else {
				allInvolvedTblConfigs[tblName] = cfg
			}
		}
	}
	return allInvolvedTblConfigs, nil
}

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
				log.L().Error("generate INSERT SQL error", zap.Error(err))
				continue
			}
			_, err = tx.ExecContext(ctx, sql)
			if err != nil {
				log.L().Error("execute SQL error", zap.Error(err))
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

func (s *DBSimulator) LoadMCP(ctx context.Context) error {
	allInvolvedTblConfigs, err := s.getAllInvolvedTableConfigs()
	if err != nil {
		return errors.Annotate(err, "load MCP error")
	}
	for tblName, tblConf := range allInvolvedTblConfigs {
		sqlGen := sqlgen.NewSQLGeneratorImpl(tblConf)
		sql, colMetas, err := sqlGen.GenLoadUniqueKeySQL()
		if err != nil {
			return errors.Annotate(err, "generate load unique key SQL error")
		}
		rows, err := s.db.QueryContext(ctx, sql)
		if err != nil {
			return errors.Annotate(err, "execute load Unique SQL error")
		}
		mcp := sqlgen.NewModificationCandidatePool()
		for rows.Next() {
			values := make([]interface{}, 0)
			for _, colMeta := range colMetas {
				valHolder := newColValueHolder(colMeta)
				if valHolder == nil {
					log.L().Error("unsupported data type",
						zap.String("column_name", colMeta.ColumnName),
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
				ukValue[colMeta.ColumnName] = getValueHolderValue(v)
			}
			theUK := &sqlgen.UniqueKey{
				RowID: -1,
				Value: ukValue,
			}
			mcp.AddUK(theUK)
			log.L().Debug("add UK value to the pool", zap.Any("uk", theUK))
		}
		if rows.Err() != nil {
			return errors.Annotate(err, "fetch rows has error")
		}
		s.mcpMap[tblName] = mcp
	}
	return nil
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

func (s *DBSimulator) StartSimulation(ctx context.Context) error {
	if s.isRunning.Load() {
		log.L().Info("the DB simulator has already been started")
		return nil
	}
	return func() error {
		s.Lock()
		defer s.Unlock()
		s.ctx, s.cancel = context.WithCancel(ctx)
		workerCount := 4
		s.wg.Add(workerCount)
		s.workerCh = make(chan WorkloadSimulator, workerCount)
		for i := 0; i < workerCount; i++ {
			go func() {
				defer s.wg.Done()
				s.workerFn(s.ctx, s.workerCh)
				log.L().Info("worker exit")
			}()
		}
		s.wg.Add(1)
		go func(ctx context.Context) {
			defer s.wg.Done()
			for s.isRunning.Load() {
				select {
				case <-ctx.Done():
					log.L().Info("context is done")
					return
				default:
					s.DoSimulation(ctx)
				}
			}
		}(s.ctx)
		s.isRunning.Store(true)
		log.L().Info("the DB simulator has been started")
		return nil
	}()
}

func (s *DBSimulator) StopSimulation() error {
	if !s.isRunning.Load() {
		log.L().Info("the server has already been closed")
		return nil
	}
	//atomic operations on closing the server
	log.L().Info("begin to stop the DB simulator")
	func() {
		s.Lock()
		defer s.Unlock()
		s.cancel()
		log.L().Info("begin to wait all the goroutines to finish")
		s.wg.Wait() //wait all sub-goroutines finished
		log.L().Info("all the goroutines finished")
		s.isRunning.Store(false)
		log.L().Info("the DB simulator is stopped")
	}()
	return nil
}

func (s *DBSimulator) DoSimulation(ctx context.Context) {
	var theWorkload WorkloadSimulator
	for {
		select {
		case <-ctx.Done():
			log.L().Info("context expired, simulation terminated")
			return
		default:
			theWorkload = func() WorkloadSimulator {
				s.workloadLock.RLock()
				defer s.workloadLock.RUnlock()
				weightMap := make(map[string]int)
				for workloadName := range s.workloadSimulators {
					weightMap[workloadName] = 1
				}
				workloadName := utils.RandomChooseKeyByWeights(weightMap)
				return s.workloadSimulators[workloadName]
			}()
		}
		select {
		case s.workerCh <- theWorkload:
			//continue on
		case <-time.After(1 * time.Second):
			//continue on
		}
	}
}
