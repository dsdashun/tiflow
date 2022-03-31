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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pingcap/errors"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/core"
	"github.com/pingcap/tiflow/dm/simulator/internal/server"
	"github.com/pingcap/tiflow/dm/simulator/internal/workload"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
)

func main() {
	var gerr error
	if err := plog.InitLogger(&plog.Config{}); err != nil {
		log.Fatalf("init logger error: %v\n", err)
	}
	defer func() {
		if err := plog.L().Sync(); err != nil {
			log.Println("sync log failed", err)
		}
		if gerr != nil {
			os.Exit(1)
		}
	}()
	cliConfig := config.NewCLIConfig()
	flag.Parse()
	if cliConfig.IsHelp {
		flag.Usage()
		return
	}
	if len(cliConfig.ConfigFile) == 0 {
		errMsg := "config file is empty"
		fmt.Fprintln(os.Stderr, errMsg)
		flag.Usage()
		gerr = errors.New(errMsg)
		return
	}

	theConfig, err := config.NewConfigFromFile(cliConfig.ConfigFile)
	if err != nil {
		gerr = errors.Annotate(err, "new config from file error")
		fmt.Fprintln(os.Stderr, gerr)
		return
	}
	// this context is for the main function context, sending signals will cancel the context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// capture the signal and handle
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		sig := <-sc
		plog.L().Info("got signal to exit", zap.Stringer("signal", sig))
		cancel()
	}()

	if len(theConfig.DataSources) == 0 {
		gerr = errors.New("no data source provided")
		plog.L().Error(gerr.Error())
		return
	}
	srv := server.NewServer()
	totalTableConfigMap := make(map[string]map[string]*config.TableConfig)
	dbSimulatorMap := make(map[string]*core.DBSimulator)
	for _, dbConfig := range theConfig.DataSources {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbConfig.UserName, dbConfig.Password, dbConfig.Host, dbConfig.Port))
		if err != nil {
			plog.L().Error("open testing DB failed", zap.Error(err))
			gerr = err
			return
		}
		if len(dbConfig.Tables) == 0 {
			gerr = errors.New("no simulating data table provided")
			plog.L().Error(gerr.Error())
			return
		}
		tblConfigMap := make(map[string]*config.TableConfig)
		for _, tblConfig := range dbConfig.Tables {
			tblConfigMap[tblConfig.TableID] = tblConfig
		}
		totalTableConfigMap[dbConfig.DataSourceID] = tblConfigMap
		theSimulator := core.NewDBSimulator(db, tblConfigMap)
		srv.SetDBSimulator(dbConfig.DataSourceID, theSimulator)
		dbSimulatorMap[dbConfig.DataSourceID] = theSimulator
	}
	plog.L().Info("begin to register workloads")
	for i, workloadConf := range theConfig.Workloads {
		plog.L().Info("add the workload into simulator")
		for _, dataSourceID := range workloadConf.DataSources {
			tblConfigMap, ok := totalTableConfigMap[dataSourceID]
			if !ok {
				errMsg := "cannot find the table config map"
				plog.L().Error(errMsg, zap.String("data source ID", dataSourceID))
				gerr = errors.New(errMsg)
				return
			}
			theSimulator := dbSimulatorMap[dataSourceID]
			workloadSimu, err := workload.NewWorkloadSimulatorImpl(tblConfigMap, workloadConf.WorkloadCode)
			if err != nil {
				gerr = errors.Annotate(err, "new workload simulator error")
				plog.L().Error(gerr.Error())
				return
			}
			theSimulator.AddWorkload(fmt.Sprintf("workload%d", i), workloadSimu)
		}
	}
	plog.L().Info("registering workloads [DONE]")
	plog.L().Info("begin to load all related table schemas")
	for _, dbConfig := range theConfig.DataSources {
		theSimulator := dbSimulatorMap[dbConfig.DataSourceID]
		if err := theSimulator.LoadAllTableSchemas(context.Background()); err != nil {
			plog.L().Error("load all table schemas error", zap.Error(err))
			gerr = err
			return
		}
	}
	plog.L().Info("loading all related table schemas [DONE]")
	if err := srv.Start(ctx); err != nil {
		plog.L().Error("start server error", zap.Error(err))
		gerr = err
		return
	}
	<-ctx.Done()
	if err := srv.Stop(); err != nil {
		plog.L().Error("stop server error", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("main exit")
}
