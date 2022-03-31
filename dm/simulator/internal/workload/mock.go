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

package workload

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
	"go.uber.org/atomic"
)

func packToString(cfgMap map[string]*config.TableConfig) (string, error) {
	tblCfgMd5s := make([][2]string, 0)
	for tblID, tblConfig := range cfgMap {
		clonedCfg := tblConfig.SortedClone()
		md5Hash := md5.New()
		jsonEnc := json.NewEncoder(md5Hash)
		if err := jsonEnc.Encode(clonedCfg); err != nil {
			return "", err
		}
		cfgHashStr := fmt.Sprintf("%x", md5Hash.Sum(nil))
		tblCfgMd5s = append(tblCfgMd5s, [2]string{tblID, cfgHashStr})
	}
	sort.Slice(tblCfgMd5s, func(i, j int) bool {
		return tblCfgMd5s[i][0] < tblCfgMd5s[j][0]
	})
	md5PairStrs := make([]string, 0)
	for _, md5Pair := range tblCfgMd5s {
		md5PairStrs = append(md5PairStrs, fmt.Sprintf("%s=%s", md5Pair[0], md5Pair[1]))
	}
	return strings.Join(md5PairStrs, ","), nil
}

// MockWorkload defines a fake workload for unit tests.
// It implements the `Simulator` interface.
type MockWorkload struct {
	sync.RWMutex
	involvedTables   map[string]struct{}
	totalExecutedMap map[string]*atomic.Uint64
	tblConfigMap     map[string]*config.TableConfig
	isEnabled        *atomic.Bool
	currentSchemaStr string
}

// NewMockWorkload creates a new `MockWorkload`.
func NewMockWorkload(tblCfgMap map[string]*config.TableConfig) (*MockWorkload, error) {
	involvedTables := make(map[string]struct{})
	for tableID := range tblCfgMap {
		involvedTables[tableID] = struct{}{}
	}
	schemaStr, err := packToString(tblCfgMap)
	if err != nil {
		return nil, err
	}
	totalExecMap := make(map[string]*atomic.Uint64)
	totalExecMap[schemaStr] = atomic.NewUint64(0)
	return &MockWorkload{
		involvedTables:   involvedTables,
		totalExecutedMap: totalExecMap,
		tblConfigMap:     tblCfgMap,
		isEnabled:        atomic.NewBool(true),
		currentSchemaStr: schemaStr,
	}, nil
}

// GetCurrentSchemaSignature gets the schema signature for the mock workload.
// It is used to do fast comparations on unit tests.
func (w *MockWorkload) GetCurrentSchemaSignature() string {
	return w.currentSchemaStr
}

// TotalExecuted gets total executed queries on the specific table schema.
// It is used in the unit tests to verify the execution stats after the schema has changed.
func (w *MockWorkload) TotalExecuted(schemaSig string) uint64 {
	w.RLock()
	defer w.RUnlock()
	auint, ok := w.totalExecutedMap[schemaSig]
	if !ok {
		return 0
	}
	return auint.Load()
}

// SimulateTrx simulates a transaction from the workload simulator.
// It implements the `Simulator` interface.
func (w *MockWorkload) SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*mcp.ModificationCandidatePool) error {
	w.RLock()
	defer w.RUnlock()
	if !w.IsEnabled() {
		return nil
	}
	w.totalExecutedMap[w.currentSchemaStr].Inc()
	return nil
}

// GetInvolvedTables collects all the involved table names in the workload.
// It implements the `Simulator` interface.
func (w *MockWorkload) GetInvolvedTables() []string {
	involvedTblStrs := make([]string, 0)
	for tblID := range w.involvedTables {
		involvedTblStrs = append(involvedTblStrs, tblID)
	}
	return involvedTblStrs
}

// SetTableConfig sets the table config of a table ID.
// It implements the `Simulator` interface.
func (w *MockWorkload) SetTableConfig(tableID string, tblConfig *config.TableConfig) error {
	if !w.DoesInvolveTable(tableID) {
		return nil
	}
	w.Lock()
	defer w.Unlock()
	w.tblConfigMap[tableID] = tblConfig
	newSchemaStr, err := packToString(w.tblConfigMap)
	if err != nil {
		return err
	}
	if _, ok := w.totalExecutedMap[newSchemaStr]; !ok {
		w.totalExecutedMap[newSchemaStr] = atomic.NewUint64(0)
	}
	w.currentSchemaStr = newSchemaStr
	return nil
}

// Enable enables this workload.
// It implements the `Simulator` interface.
func (w *MockWorkload) Enable() {
	w.isEnabled.Store(true)
}

// Disable disables this workload.
// It implements the `Simulator` interface.
func (w *MockWorkload) Disable() {
	w.isEnabled.Store(false)
}

// IsEnabled checks whether this workload is enabled or not.
// It implements the `Simulator` interface.
func (w *MockWorkload) IsEnabled() bool {
	return w.isEnabled.Load()
}

// DoesInvolveTable checks whether this workload involves the specified table.
// It implements the `Simulator` interface.
func (w *MockWorkload) DoesInvolveTable(tableID string) bool {
	_, ok := w.involvedTables[tableID]
	return ok
}
