package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chaos-mesh/go-sqlsmith/types"
	"go.uber.org/zap"

	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/core"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

func main() {
	var (
		err  error
		gerr error
	)
	defer func() {
		err = plog.L().Sync()
		if err != nil {
			log.Println("sync log failed", err)
		}
		if gerr != nil {
			os.Exit(1)
		}
	}()
	plog.InitLogger(&plog.Config{})
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

	tableInfo := &types.Table{
		DB:    "games",
		Table: "members",
		Type:  "BASE TABLE",
		Columns: map[string]*types.Column{
			"id": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "id",
				DataType: "int",
				DataLen:  11,
			},
			"name": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "name",
				DataType: "varchar",
				DataLen:  255,
			},
			"age": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "age",
				DataType: "int",
				DataLen:  11,
			},
			"team_id": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "team_id",
				DataType: "int",
				DataLen:  11,
			},
		},
	}
	ukColumns := map[string]*types.Column{
		"id": tableInfo.Columns["id"],
	}

	db, err := sql.Open("mysql", "root:guanliyuanmima@tcp(127.0.0.1:13306)/games")
	if err != nil {
		plog.L().Error("open testing DB failed", zap.Error(err))
		gerr = err
		return
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(tableInfo, ukColumns)
	theSimulator := core.NewSimulatorImpl(db, sqlGen)

	plog.L().Info("begin to prepare table data")
	err = theSimulator.PrepareData(context.Background(), 4096)
	if err != nil {
		plog.L().Error("prepare table data failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("prepare table data [DONE]")
	plog.L().Info("begin to load UKs into MCP")
	err = theSimulator.LoadMCP(context.Background())
	if err != nil {
		plog.L().Error("load UKs of table into MCP failed", zap.Error(err))
		gerr = err
		return
	}
	plog.L().Info("loading UKs into MCP [DONE]")
	plog.L().Info("start simulation")
	theSimulator.DoSimulation(ctx)
	<-ctx.Done()
	plog.L().Info("simulation terminated")
	plog.L().Info("main exit")
}
