package core

import (
	"context"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/utils"
)

func workerFn(ctx context.Context, workloadCh <-chan WorkloadSimulator) {
	for {
		select {
		case <-ctx.Done():
			log.L().Info("worker context is terminated")
			return
		case theWorkload := <-workloadCh:
			err := theWorkload.SimulateTrx(ctx)
			if err != nil {
				log.L().Error("simulate a trx error", zap.Error(err))
			}
		case <-time.After(100 * time.Millisecond):
			//continue on
		}
	}
}

type DBSimulator struct {
	sync.RWMutex
	isRunning          atomic.Bool
	wg                 sync.WaitGroup
	workloadLock       sync.RWMutex
	workloadSimulators map[string]WorkloadSimulator
	ctx                context.Context
	cancel             func()
	workerCh           chan WorkloadSimulator
}

func NewDBSimulator() *DBSimulator {
	return &DBSimulator{
		workloadSimulators: make(map[string]WorkloadSimulator),
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
				workerFn(s.ctx, s.workerCh)
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
				for tableName := range s.workloadSimulators {
					weightMap[tableName] = 1
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
