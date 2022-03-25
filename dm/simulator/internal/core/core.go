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

// Package core contains the core logic of a simulator.
// It includes the logic to simulate a single workload as well as the framework for a simulator.
package core

import (
	"context"
	"database/sql"

	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
)

type MCPLoader interface {
	LoadMCP(ctx context.Context, tblConf *config.TableConfig) (*mcp.ModificationCandidatePool, error)
}

// Simulator defines all the basic operations of a simulator.
// In the future, each method will be mapped to an API
type DBSimulatorInterface interface {
	// StartSimulation starts the simulation.
	StartSimulation(ctx context.Context) error

	// StopSimulation stops the simulation.
	StopSimulation() error

	// GetTableConfig gets the table config of a table
	GetTableConfig(tableName string) *config.TableConfig

	// Prepare prepares for starting the simulation.
	Prepare(ctx context.Context) error

	// GetDB gets the *sql.DB object for the simulator.
	GetDB() *sql.DB

	// GetContext gets the context of this simulator
	GetContext() context.Context
}
