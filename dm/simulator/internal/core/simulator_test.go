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
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/schema"
	"github.com/pingcap/tiflow/dm/simulator/internal/workload"
)

type testDBSimulatorSuite struct {
	suite.Suite
}

func (s *testDBSimulatorSuite) SetupSuite() {
	s.Require().Nil(log.InitLogger(&log.Config{}))
}

func mockPrepareData(mock sqlmock.Sqlmock, recordCount int) {
	mock.ExpectBegin()
	mock.ExpectExec("^TRUNCATE TABLE (.+)").WillReturnResult(sqlmock.NewResult(0, 999))
	for i := 0; i < recordCount; i++ {
		mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()
}

func (s *testDBSimulatorSuite) TestPrepareMCP() {
	tableConfigMap := map[string]*config.TableConfig{
		"members": config.NewTemplateTableConfigForTest(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	recordCount := 128
	theSimulator := NewDBSimulator(db, tableConfigMap)
	w1, err := workload.NewMockWorkload(tableConfigMap)
	s.Require().Nil(err)
	theSimulator.AddWorkload("dummy_members", w1)
	mockPrepareData(mock, recordCount)
	s.Require().Nil(theSimulator.PrepareData(ctx, recordCount))
}

func (s *testDBSimulatorSuite) TestChooseWorkload() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg1 := config.NewTemplateTableConfigForTest()
	cfg1.TableID = "tbl01"
	cfg1.TableName = "member01"
	cfg2 := config.NewTemplateTableConfigForTest()
	cfg2.TableID = "tbl02"
	cfg2.TableName = "member02"
	simu := NewDBSimulator(nil, map[string]*config.TableConfig{
		cfg1.TableID: cfg1,
		cfg2.TableID: cfg2,
	})
	w1, err := workload.NewMockWorkload(map[string]*config.TableConfig{
		cfg1.TableID: cfg1,
	})
	s.Require().Nil(err)
	w2, err := workload.NewMockWorkload(map[string]*config.TableConfig{
		cfg2.TableID: cfg2,
	})
	s.Require().Nil(err)
	w3, err := workload.NewMockWorkload(map[string]*config.TableConfig{
		cfg1.TableID: cfg1,
		cfg2.TableID: cfg2,
	})
	s.Require().Nil(err)
	simu.AddWorkload("w1", w1)
	simu.AddWorkload("w2", w2)
	simu.AddWorkload("w3", w3)
	weightMap := make(map[string]int)
	for tableName := range simu.workloadSimulators {
		weightMap[tableName] = 1
	}
	for i := 0; i < 100; i++ {
		workloadName := randomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		assert.Nil(s.T(), theWorkload.SimulateTrx(ctx, nil, nil))
	}
	w1CurrentExecuted := w1.TotalExecuted(w1.GetCurrentSchemaSignature())
	w2CurrentExecuted := w2.TotalExecuted(w2.GetCurrentSchemaSignature())
	w3CurrentExecuted := w3.TotalExecuted(w3.GetCurrentSchemaSignature())
	s.Greater(w1CurrentExecuted, uint64(0), "workload 01 should at least execute once")
	s.Greater(w2CurrentExecuted, uint64(0), "workload 02 should at least execute once")
	s.Greater(w3CurrentExecuted, uint64(0), "workload 03 should at least execute once")
	w2.Disable()
	for i := 0; i < 100; i++ {
		workloadName := randomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		if !theWorkload.IsEnabled() {
			continue
		}
		s.Nil(theWorkload.SimulateTrx(ctx, nil, nil))
	}
	s.Equal(w2CurrentExecuted, w2.TotalExecuted(w2.GetCurrentSchemaSignature()), "workload 02 should not be executed after disabled")
	w1CurrentExecuted = w1.TotalExecuted(w1.GetCurrentSchemaSignature())
	w3CurrentExecuted = w3.TotalExecuted(w3.GetCurrentSchemaSignature())
	w2.Enable()
	simu.RemoveWorkload("w3")
	weightMap = make(map[string]int)
	for tableName := range simu.workloadSimulators {
		weightMap[tableName] = 1
	}
	for i := 0; i < 100; i++ {
		workloadName := randomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		if !theWorkload.IsEnabled() {
			continue
		}
		s.Nil(theWorkload.SimulateTrx(ctx, nil, nil))
	}
	assert.Greater(s.T(), w1.TotalExecuted(w1.GetCurrentSchemaSignature()), w1CurrentExecuted, "workload 01 should at least execute once")
	assert.Greater(s.T(), w2.TotalExecuted(w2.GetCurrentSchemaSignature()), w2CurrentExecuted, "workload 02 should at least execute once")
	assert.Equal(s.T(), w3.TotalExecuted(w3.GetCurrentSchemaSignature()), w3CurrentExecuted, "workload 03 should keep the executed count")
}

func (s *testDBSimulatorSuite) TestStartStopSimulation() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg1 := config.NewTemplateTableConfigForTest()
	cfg1.TableID = "tbl01"
	cfg1.TableName = "member01"
	cfg2 := config.NewTemplateTableConfigForTest()
	cfg2.TableID = "tbl02"
	cfg2.TableName = "member02"
	cfgMap := map[string]*config.TableConfig{
		cfg1.TableID: cfg1,
		cfg2.TableID: cfg2,
	}
	mockSG := schema.NewMockSchemaGetter()
	mockSG.SetFromTableConfig(cfg1)
	mockSG.SetFromTableConfig(cfg2)

	mockMCPLoader := NewMockMCPLoader(4096)

	simu := NewDBSimulator(nil, cfgMap)
	simu.sg = mockSG
	simu.mcpLoader = mockMCPLoader
	w1, err := workload.NewMockWorkload(map[string]*config.TableConfig{
		cfg1.TableID: cfg1,
	})
	s.Require().Nil(err)
	w2, err := workload.NewMockWorkload(map[string]*config.TableConfig{
		cfg2.TableID: cfg2,
	})
	s.Require().Nil(err)
	w3, err := workload.NewMockWorkload(map[string]*config.TableConfig{
		cfg1.TableID: cfg1,
		cfg2.TableID: cfg2,
	})
	s.Require().Nil(err)
	simu.AddWorkload("w1", w1)
	simu.AddWorkload("w2", w2)
	simu.AddWorkload("w3", w3)
	s.Require().Nil(simu.StartSimulation(ctx))
	time.Sleep(1 * time.Second)
	s.Require().Nil(simu.StopSimulation())
	w1SchemaSig := w1.GetCurrentSchemaSignature()
	w2SchemaSig := w2.GetCurrentSchemaSignature()
	w3SchemaSig := w3.GetCurrentSchemaSignature()
	s.Greater(w1.TotalExecuted(w1SchemaSig), uint64(0), "workload 01 should at least execute once")
	s.Greater(w2.TotalExecuted(w2SchemaSig), uint64(0), "workload 02 should at least execute once")
	s.Greater(w3.TotalExecuted(w3SchemaSig), uint64(0), "workload 03 should at least execute once")
	// start again
	newCfg2 := cfg2.SortedClone()
	newCfg2.Columns = append(newCfg2.Columns, &config.ColumnDefinition{
		ColumnName: "dummycol",
		DataType:   "varchar",
	})
	s.Require().Nil(simu.StartSimulation(ctx))
	mockSG.SetFromTableConfig(newCfg2)
	time.Sleep(1 * time.Second)
	s.Require().Nil(simu.StopSimulation())

	newW1SchemaSig := w1.GetCurrentSchemaSignature()
	newW2SchemaSig := w2.GetCurrentSchemaSignature()
	newW3SchemaSig := w3.GetCurrentSchemaSignature()

	s.Equal(w1SchemaSig, newW1SchemaSig)
	s.NotEqual(w2SchemaSig, newW2SchemaSig)
	s.NotEqual(w3SchemaSig, newW3SchemaSig)
	s.Greater(w1.TotalExecuted(newW1SchemaSig), uint64(0), "workload 01 should at least execute once")
	s.Greater(w2.TotalExecuted(newW2SchemaSig), uint64(0), "workload 02 should at least execute once")
	s.Greater(w3.TotalExecuted(newW3SchemaSig), uint64(0), "workload 03 should at least execute once")
}

func TestDBSimulatorSuite(t *testing.T) {
	suite.Run(t, &testDBSimulatorSuite{})
}
