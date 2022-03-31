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

package schema

import (
	"context"
	"errors"
	"fmt"
	"sync"

	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"go.uber.org/zap"
)

// MockSchemaGetter is a `Getter` implementation.
// It is used in unit tests.
type MockSchemaGetter struct {
	sync.RWMutex
	columnDefs map[string][]*config.ColumnDefinition
	ukColNames map[string][]string
}

// NewMockSchemaGetter generates a new `MockSchemaGetter`.
func NewMockSchemaGetter() *MockSchemaGetter {
	return &MockSchemaGetter{
		columnDefs: make(map[string][]*config.ColumnDefinition),
		ukColNames: make(map[string][]string),
	}
}

// GetColumnDefinitions gets the column definitions of a table.
// It implements the `Getter` interface.
func (g *MockSchemaGetter) GetColumnDefinitions(ctx context.Context, dbName string, tableName string) ([]*config.ColumnDefinition, error) {
	g.RLock()
	defer g.RUnlock()
	keyName := fmt.Sprintf("%s.%s", dbName, tableName)
	colDefs, ok := g.columnDefs[keyName]
	if !ok {
		errMsg := "cannot find column definition"
		plog.L().Error(errMsg, zap.String("db_name", dbName), zap.String("table_name", tableName))
		return nil, errors.New(errMsg)
	}
	clonedColDefs := make([]*config.ColumnDefinition, 0)
	for _, colDef := range colDefs {
		clonedColDefs = append(clonedColDefs, &config.ColumnDefinition{
			ColumnName: colDef.ColumnName,
			DataType:   colDef.DataType,
		})
	}
	return clonedColDefs, nil
}

// GetUniqueKeyColumns gets the columns of a unique key in a table.
// It implements the `Getter` interface.
func (g *MockSchemaGetter) GetUniqueKeyColumns(ctx context.Context, dbName string, tableName string) ([]string, error) {
	g.RLock()
	defer g.RUnlock()
	keyName := fmt.Sprintf("%s.%s", dbName, tableName)
	ukColNames, ok := g.ukColNames[keyName]
	if !ok {
		errMsg := "cannot find uk column names"
		plog.L().Error(errMsg, zap.String("db_name", dbName), zap.String("table_name", tableName))
		return nil, errors.New(errMsg)
	}
	return append([]string{}, ukColNames...), nil
}

// SetFromTableConfig sets the returned table schemas from a table config.
func (g *MockSchemaGetter) SetFromTableConfig(tblConfig *config.TableConfig) {
	clonedColDefs := config.CloneSortedColumnDefinitions(tblConfig.Columns)
	clonedUKCols := append([]string{}, tblConfig.UniqueKeyColumnNames...)
	keyName := fmt.Sprintf("%s.%s", tblConfig.DatabaseName, tblConfig.TableName)
	g.Lock()
	defer g.Unlock()
	g.columnDefs[keyName] = clonedColDefs
	g.ukColNames[keyName] = clonedUKCols
}
