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
	var tblCfgMd5s [][2]string
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
	var md5PairStrs []string
	for _, md5Pair := range tblCfgMd5s {
		md5PairStrs = append(md5PairStrs, fmt.Sprintf("%s=%s", md5Pair[0], md5Pair[1]))
	}
	return strings.Join(md5PairStrs, ","), nil
}

type MockWorkload struct {
	sync.RWMutex
	involvedTables   map[string]struct{}
	totalExecutedMap map[string]*atomic.Uint64
	tblConfigMap     map[string]*config.TableConfig
	isEnabled        *atomic.Bool
	currentSchemaStr string
}

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

func (w *MockWorkload) GetCurrentSchemaSignature() string {
	return w.currentSchemaStr
}

func (w *MockWorkload) TotalExecuted(schemaSig string) uint64 {
	w.RLock()
	defer w.RUnlock()
	auint, ok := w.totalExecutedMap[schemaSig]
	if !ok {
		return 0
	}
	return auint.Load()
}

func (w *MockWorkload) SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*mcp.ModificationCandidatePool) error {
	w.RLock()
	defer w.RUnlock()
	if !w.IsEnabled() {
		return nil
	}
	w.totalExecutedMap[w.currentSchemaStr].Inc()
	return nil
}

func (w *MockWorkload) GetInvolvedTables() []string {
	var involvedTblStrs []string
	for tblID := range w.involvedTables {
		involvedTblStrs = append(involvedTblStrs, tblID)
	}
	return involvedTblStrs
}

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

func (w *MockWorkload) Enable() {
	w.isEnabled.Store(true)
}

func (w *MockWorkload) Disable() {
	w.isEnabled.Store(false)
}

func (w *MockWorkload) IsEnabled() bool {
	return w.isEnabled.Load()
}

func (w *MockWorkload) DoesInvolveTable(tableID string) bool {
	_, ok := w.involvedTables[tableID]
	return ok
}
