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
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

func main() {
	var (
		err  error
		gerr error
	)
	plog.InitLogger(&plog.Config{})
	defer func() {
		err = plog.L().Sync()
		if err != nil {
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
		os.Exit(0)
	}
	if len(cliConfig.ConfigFile) == 0 {
		fmt.Fprintln(os.Stderr, "config file is empty")
		flag.Usage()
		os.Exit(1)
	}

	theConfig, err := config.NewConfigFromFile(cliConfig.ConfigFile)
	if err != nil {
		log.Fatalf("new config from file error: %v\n", err)
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
	sqlGen := sqlgen.NewSQLGeneratorImpl(dbConfig.Tables[0])
	tableSimu := core.NewWorkloadSimulatorImpl(db, sqlGen)
	theSimulator := core.NewDBSimulator()

	plog.L().Info("begin to prepare table data")
	err = tableSimu.PrepareData(context.Background(), 4096)
	if err != nil {
		plog.L().Error("prepare table data failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("prepare table data [DONE]")
	plog.L().Info("begin to load UKs into MCP")
	err = tableSimu.LoadMCP(context.Background())
	if err != nil {
		plog.L().Error("load UKs of table into MCP failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("loading UKs into MCP [DONE]")

	plog.L().Info("add the workload into simulator")
	theSimulator.AddWorkload("games.members/RANDOM", tableSimu)

	plog.L().Info("start simulation")
	err = theSimulator.StartSimulation(ctx)
	if err != nil {
		plog.L().Error("start simulation failed", zap.Error(err))
		gerr = err
		return
	}
	<-ctx.Done()
	plog.L().Info("simulation terminated")
	err = theSimulator.StopSimulation()
	if err != nil {
		plog.L().Error("stop simulation failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("main exit")
}
