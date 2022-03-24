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

// Package workload contains the execution logic of a simulator workload.
package workload

import (
	"context"
	"database/sql"

	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
)

// WorkloadSimulator defines all the basic operations for simulating a transaction of a workload.
type WorkloadSimulator interface {
	// SimulateTrx simulates a transaction from the workload simulator.
	SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*mcp.ModificationCandidatePool) error

	// GetInvolvedTables collects all the involved table names in the workload.
	// This operation is often used when the caller wants to collect all the involved tables for several workloads,
	// so that only the tables really needed are preparred for data.
	GetInvolvedTables() []string
	// SetTableConfig sets the table config of a table ID.
	// It is called after the table structure has changed,
	// so that the workload simulator should simulate transactions
	// using the latest table structure.
	SetTableConfig(tableID string, tblConfig *config.TableConfig) error
	// Enable enables this workload.
	Enable()
	// Disable disables this workload.
	// A disabled workload won't be executed unitil it is enabled again.
	Disable()
	// IsEnabled checks whether this workload is enabled or not.
	IsEnabled() bool
	// DoesInvolveTable checks whether this workload involves the specified table.
	DoesInvolveTable(tableID string) bool
}
