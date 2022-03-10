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
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/core"
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
	dbConfig := theConfig.DataSources[0]
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

	theSimulator := core.NewDBSimulator(db, tblConfigMap)
	for i, workloadConf := range theConfig.Workloads {
		workloadSimu, err := core.NewWorkloadSimulatorImpl(tblConfigMap, workloadConf.WorkloadCode)
		if err != nil {
			gerr = errors.Annotate(err, "new workload simulator error")
			plog.L().Error(gerr.Error())
			return
		}

		plog.L().Info("add the workload into simulator")
		theSimulator.AddWorkload(fmt.Sprintf("workload%d", i), workloadSimu)
	}
	plog.L().Info("begin to prepare table data")
	if err := theSimulator.PrepareData(context.Background(), 4096); err != nil {
		plog.L().Error("prepare table data failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("prepare table data [DONE]")
	plog.L().Info("begin to load UKs into MCP")
	if err := theSimulator.LoadMCP(context.Background()); err != nil {
		plog.L().Error("load UKs of table into MCP failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("loading UKs into MCP [DONE]")

	plog.L().Info("start simulation")
	if err := theSimulator.StartSimulation(ctx); err != nil {
		plog.L().Error("start simulation failed", zap.Error(err))
		gerr = err
		return
	}
	<-ctx.Done()
	plog.L().Info("simulation terminated")
	if err := theSimulator.StopSimulation(); err != nil {
		plog.L().Error("stop simulation failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("main exit")
}
