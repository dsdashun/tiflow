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

	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
)

// Simulator defines all the basic operations of a simulator.
type Simulator interface {
	// StartSimulation starts the simulation.
	StartSimulation(ctx context.Context) error

	// StopSimulation stops the simulation.
	StopSimulation() error
}

// WorkloadSimulator defines all the basic operations for simulating a transaction of a workload.
type WorkloadSimulator interface {
	// SimulateTrx simulates a transaction from the workload simulator.
	SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*mcp.ModificationCandidatePool) error

	// GetInvolvedTables collects all the involved table names in the workload.
	// This operation is often used when the caller wants to collect all the involved tables for several workloads,
	// so that only the tables really needed are preparred for data.
	GetInvolvedTables() []string
}
