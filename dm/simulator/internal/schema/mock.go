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

type MockSchemaGetter struct {
	sync.RWMutex
	columnDefs map[string][]*config.ColumnDefinition
	ukColNames map[string][]string
}

func NewMockSchemaGetter() *MockSchemaGetter {
	return &MockSchemaGetter{
		columnDefs: make(map[string][]*config.ColumnDefinition),
		ukColNames: make(map[string][]string),
	}
}

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
	var clonedColDefs []*config.ColumnDefinition
	for _, colDef := range colDefs {
		clonedColDefs = append(clonedColDefs, &config.ColumnDefinition{
			ColumnName: colDef.ColumnName,
			DataType:   colDef.DataType,
		})
	}
	return clonedColDefs, nil
}

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

func (g *MockSchemaGetter) SetFromTableConfig(tblConfig *config.TableConfig) {
	clonedColDefs := config.CloneSortedColumnDefinitions(tblConfig.Columns)
	clonedUKCols := append([]string{}, tblConfig.UniqueKeyColumnNames...)
	keyName := fmt.Sprintf("%s.%s", tblConfig.DatabaseName, tblConfig.TableName)
	g.Lock()
	defer g.Unlock()
	g.columnDefs[keyName] = clonedColDefs
	g.ukColNames[keyName] = clonedUKCols
}
