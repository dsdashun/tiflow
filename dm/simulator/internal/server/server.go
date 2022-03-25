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

package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/errors"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/core"
	"github.com/pingcap/tiflow/dm/simulator/internal/openapi"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Server struct {
	sync.RWMutex
	simulators map[string]core.DBSimulatorInterface
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     func()
	isRunning  atomic.Bool
	httpSrv    *http.Server
}

func NewServer() *Server {
	return &Server{
		simulators: make(map[string]core.DBSimulatorInterface),
	}
}

func (s *Server) SetDBSimulator(dbName string, simu core.DBSimulatorInterface) {
	s.Lock()
	defer s.Unlock()
	s.simulators[dbName] = simu
}

func (s *Server) GetDBSimulator(dbName string) core.DBSimulatorInterface {
	s.RLock()
	defer s.RUnlock()
	simu, ok := s.simulators[dbName]
	if !ok {
		return nil
	}
	return simu
}

func (s *Server) Start(ctx context.Context) error {
	if s.isRunning.Load() {
		// has already started
		return nil
	}
	var gerr error
	s.Lock()
	defer s.Unlock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	defer func() {
		if gerr != nil {
			if err := s.doStop(); err != nil {
				plog.L().Error("stop the server error", zap.Error(err))
			}
		}
	}()
	r := gin.Default()
	apiServer := NewAPIServer(s)
	r = openapi.RegisterHandlers(r, apiServer)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.httpSrv = &http.Server{
			Addr:    "0.0.0.0:8888",
			Handler: r,
		}
		if err := s.httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			plog.L().Error("listen and serve error", zap.Error(err))
		}
		<-s.ctx.Done()
	}()
	for dataSourceID, simu := range s.simulators {
		plog.L().Info("begin to prepare before starting simulation")
		if err := simu.Prepare(s.ctx); err != nil {
			errMsg := "prepare table data failed"
			plog.L().Error(errMsg, zap.Error(err))
			gerr = errors.Annotate(err, errMsg)
			return gerr
		}
		plog.L().Info("preparation before starting simulation [DONE]")
		plog.L().Info("start simulation", zap.String("data source", dataSourceID))
		if err := simu.StartSimulation(s.ctx); err != nil {
			errMsg := "start simulation failed"
			plog.L().Error(errMsg, zap.Error(err), zap.String("data source", dataSourceID))
			gerr = errors.Annotate(err, errMsg)
			return gerr
		}
	}
	s.isRunning.Store(true)
	return gerr
}

func (s *Server) Stop() error {
	if !s.isRunning.Load() {
		// has already stopped
		return nil
	}
	s.Lock()
	defer s.Unlock()
	if err := s.doStop(); err != nil {
		return err
	}
	s.ctx = nil
	s.cancel = nil
	s.httpSrv = nil
	s.isRunning.Store(false)
	return nil
}

func (s *Server) doStop() error {
	s.cancel()
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(stopCtx); err != nil {
		plog.L().Error("shutdown http server error", zap.Error(err))
		if err := s.httpSrv.Close(); err != nil {
			errMsg := "forcefully close http server error"
			plog.L().Error(errMsg, zap.Error(err))
			return errors.New(errMsg)
		}
	}

	for _, simu := range s.simulators {
		if err := simu.StopSimulation(); err != nil {
			errMsg := "stop simulation failed"
			plog.L().Error(errMsg, zap.Error(err))
			return errors.Annotate(err, errMsg)
		}
	}
	s.wg.Wait()
	return nil
}
