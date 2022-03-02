package core

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/utils"
)

type dummyWorkload struct {
	Name          string
	TotalExecuted uint64
}

func (w *dummyWorkload) SimulateTrx(ctx context.Context) error {
	//log.L().Info("simulated a transaction\n", zap.String("workload_name", w.Name))
	atomic.AddUint64(&w.TotalExecuted, 1)
	return nil
}

type testDBSimulatorSuite struct {
	suite.Suite
}

func (s *testDBSimulatorSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
}

func (s *testDBSimulatorSuite) TestChooseWorkload() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	simu := NewDBSimulator()
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
		theWorkload.SimulateTrx(ctx)
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
		theWorkload.SimulateTrx(ctx)
	}
	assert.Greater(s.T(), w1.TotalExecuted, w1CurrentExecuted, "workload 01 should at least execute once")
	assert.Greater(s.T(), w2.TotalExecuted, w2CurrentExecuted, "workload 02 should at least execute once")
	assert.Equal(s.T(), w3.TotalExecuted, w3CurrentExecuted, "workload 03 should keep the executed count")
}

/*
func (s *testDBSimulatorSuite) TestSimulationLoopBasic() {
	simu := NewDBSimulator()
	w1 := &dummyWorkload{
		Name: "workload01",
	}
	w2 := &dummyWorkload{
		Name: "workload02",
	}
	simu.AddWorkload("w1", w1)
	simu.AddWorkload("w2", w2)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	simu.DoSimulation(ctx)
	assert.Greater(s.T(), w1.TotalExecuted, 0, "workload 01 should at least execute once")
	assert.Greater(s.T(), w2.TotalExecuted, 0, "workload 02 should at least execute once")
}
*/

func (s *testDBSimulatorSuite) TestStartStopSimulation() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	simu := NewDBSimulator()
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
