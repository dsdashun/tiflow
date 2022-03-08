package core

import (
	"context"
	"database/sql"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
	"github.com/pingcap/tiflow/dm/simulator/internal/utils"
)

type dummyWorkload struct {
	Name          string
	TotalExecuted uint64
}

func (w *dummyWorkload) SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*sqlgen.ModificationCandidatePool) error {
	//log.L().Info("simulated a transaction\n", zap.String("workload_name", w.Name))
	atomic.AddUint64(&w.TotalExecuted, 1)
	return nil
}

func (w *dummyWorkload) GetInvolvedTables() []string {
	return []string{w.Name}
}

type testDBSimulatorSuite struct {
	suite.Suite
}

func (s *testDBSimulatorSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
}

func mockPrepareData(mock sqlmock.Sqlmock, recordCount int) {
	mock.ExpectBegin()
	mock.ExpectExec("^TRUNCATE TABLE (.+)").WillReturnResult(sqlmock.NewResult(0, 999))
	for i := 0; i < recordCount; i++ {
		mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()
}

func mockLoadUKs(mock sqlmock.Sqlmock, recordCount int) {
	expectRows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < recordCount; i++ {
		expectRows.AddRow(rand.Int())
	}
	mock.ExpectQuery("^SELECT").WillReturnRows(expectRows)
}

func (s *testDBSimulatorSuite) TestPrepareMCP() {
	tableConfigMap := map[string]*config.TableConfig{
		"members": &config.TableConfig{
			TableID:      "members",
			DatabaseName: "games",
			TableName:    "members",
			Columns: []*config.ColumnDefinition{
				&config.ColumnDefinition{
					ColumnName: "id",
					DataType:   "int",
					DataLen:    11,
				},
				&config.ColumnDefinition{
					ColumnName: "name",
					DataType:   "varchar",
					DataLen:    255,
				},
				&config.ColumnDefinition{
					ColumnName: "age",
					DataType:   "int",
					DataLen:    11,
				},
				&config.ColumnDefinition{
					ColumnName: "team_id",
					DataType:   "int",
					DataLen:    11,
				},
			},
			UniqueKeyColumnNames: []string{"id"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	recordCount := 128
	theSimulator := NewDBSimulator(db, tableConfigMap, WithPrepareRecordCount(recordCount))
	w1 := &dummyWorkload{
		Name: "members",
	}
	theSimulator.AddWorkload("dummy_members", w1)
	mockPrepareData(mock, recordCount)
	err = theSimulator.PrepareData(ctx, recordCount)
	assert.Nil(s.T(), err)
	mockLoadUKs(mock, recordCount)
	err = theSimulator.LoadMCP(ctx)
	assert.Nil(s.T(), err)
	assert.Equalf(s.T(), recordCount, theSimulator.mcpMap["members"].Len(), "the mcp should have %d items", recordCount)
}

func (s *testDBSimulatorSuite) TestChooseWorkload() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	simu := NewDBSimulator(nil, nil)
	w1 := &dummyWorkload{
		Name: "workload01",
	}
	w2 := &dummyWorkload{
		Name: "workload02",
	}
	w3 := &dummyWorkload{
		Name: "workload03",
	}
	simu.AddWorkload("w1", w1)
	simu.AddWorkload("w2", w2)
	simu.AddWorkload("w3", w3)
	weightMap := make(map[string]int)
	for tableName := range simu.workloadSimulators {
		weightMap[tableName] = 1
	}
	for i := 0; i < 100; i++ {
		workloadName := utils.RandomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		theWorkload.SimulateTrx(ctx, nil, nil)
	}
	w1CurrentExecuted := w1.TotalExecuted
	w2CurrentExecuted := w2.TotalExecuted
	w3CurrentExecuted := w3.TotalExecuted
	assert.Greater(s.T(), w1CurrentExecuted, uint64(0), "workload 01 should at least execute once")
	assert.Greater(s.T(), w2CurrentExecuted, uint64(0), "workload 02 should at least execute once")
	assert.Greater(s.T(), w3CurrentExecuted, uint64(0), "workload 03 should at least execute once")
	simu.RemoveWorkload("w3")
	weightMap = make(map[string]int)
	for tableName := range simu.workloadSimulators {
		weightMap[tableName] = 1
	}
	for i := 0; i < 100; i++ {
		workloadName := utils.RandomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		theWorkload.SimulateTrx(ctx, nil, nil)
	}
	assert.Greater(s.T(), w1.TotalExecuted, w1CurrentExecuted, "workload 01 should at least execute once")
	assert.Greater(s.T(), w2.TotalExecuted, w2CurrentExecuted, "workload 02 should at least execute once")
	assert.Equal(s.T(), w3.TotalExecuted, w3CurrentExecuted, "workload 03 should keep the executed count")
}

func (s *testDBSimulatorSuite) TestStartStopSimulation() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	simu := NewDBSimulator(nil, nil)
	w1 := &dummyWorkload{
		Name: "workload01",
	}
	w2 := &dummyWorkload{
		Name: "workload02",
	}
	simu.AddWorkload("w1", w1)
	simu.AddWorkload("w2", w2)
	err = simu.StartSimulation(ctx)
	assert.Nil(s.T(), err)
	time.Sleep(1 * time.Second)
	err = simu.StopSimulation()
	assert.Nil(s.T(), err)
	assert.Greater(s.T(), w1.TotalExecuted, uint64(0), "workload 01 should at least execute once")
	assert.Greater(s.T(), w2.TotalExecuted, uint64(0), "workload 02 should at least execute once")
}

func TestDBSimulatorSuite(t *testing.T) {
	suite.Run(t, &testDBSimulatorSuite{})
}
