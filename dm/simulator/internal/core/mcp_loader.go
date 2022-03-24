package core

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/pingcap/errors"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
	"go.uber.org/zap"
)

type MCPLoaderImpl struct {
	db *sql.DB
}

func NewMCPLoaderImpl(db *sql.DB) *MCPLoaderImpl {
	return &MCPLoaderImpl{
		db: db,
	}
}

func (l *MCPLoaderImpl) LoadMCP(ctx context.Context, tblConf *config.TableConfig) (*mcp.ModificationCandidatePool, error) {
	sqlGen := sqlgen.NewSQLGeneratorImpl(tblConf)
	sql, colMetas, err := sqlGen.GenLoadUniqueKeySQL()
	if err != nil {
		return nil, errors.Annotate(err, "generate load unique key SQL error")
	}
	rows, err := l.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, errors.Annotate(err, "execute load Unique SQL error")
	}
	defer rows.Close()
	theMCP := mcp.NewModificationCandidatePool(8192)
	for rows.Next() {
		values := make([]interface{}, 0)
		for _, colMeta := range colMetas {
			valHolder := newColValueHolder(colMeta)
			if valHolder == nil {
				plog.L().Error("unsupported data type",
					zap.String("column_name", colMeta.ColumnName),
					zap.String("data_type", colMeta.DataType),
				)
				return nil, errors.Trace(ErrUnsupportedColumnType)
			}
			values = append(values, valHolder)
		}
		err = rows.Scan(values...)
		if err != nil {
			return nil, errors.Annotate(err, "scan values error")
		}
		ukValue := make(map[string]interface{})
		for i, v := range values {
			colMeta := colMetas[i]
			ukValue[colMeta.ColumnName] = getValueHolderValue(v)
		}
		theUK := mcp.NewUniqueKey(-1, ukValue)
		if addErr := theMCP.AddUK(theUK); addErr != nil {
			plog.L().Error("add UK into MCP error", zap.Error(addErr), zap.String("unique_key", theUK.String()))
		}
		plog.L().Debug("add UK value to the pool", zap.Any("uk", theUK))
	}
	if rows.Err() != nil {
		return nil, errors.Annotate(err, "fetch rows has error")
	}
	return theMCP, nil
}

func newColValueHolder(colMeta *config.ColumnDefinition) interface{} {
	switch colMeta.DataType {
	case "int":
		return new(int)
	case "varchar":
		return new(string)
	default:
		return nil
	}
}

func getValueHolderValue(valueHolder interface{}) interface{} {
	switch vh := valueHolder.(type) {
	case *int:
		return *vh
	case *string:
		return *vh
	default:
		return nil
	}
}

type MockMCPLoader struct {
	recordCount int
}

func NewMockMCPLoader(recordCount int) *MockMCPLoader {
	return &MockMCPLoader{
		recordCount: recordCount,
	}
}

func (l *MockMCPLoader) LoadMCP(ctx context.Context, tblConf *config.TableConfig) (*mcp.ModificationCandidatePool, error) {
	theMCP := mcp.NewModificationCandidatePool(8192)
	colDefMap := config.GenerateColumnDefinitionsMap(tblConf.Columns)
	for i := 0; i < l.recordCount; i++ {
		ukValues := make(map[string]interface{})
		for _, ukCol := range tblConf.UniqueKeyColumnNames {
			colDef, ok := colDefMap[ukCol]
			if !ok {
				errMsg := "cannot find column definition"
				plog.L().Error(errMsg, zap.String("column", ukCol))
				return nil, errors.New(errMsg)
			}
			switch colDef.DataType {
			case "int":
				ukValues[ukCol] = rand.Int()
			case "varchar":
				ukValues[ukCol] = fmt.Sprintf("STRVAL_%d", rand.Int())
			default:
				errMsg := "unsupported data type"
				plog.L().Error(errMsg, zap.String("column", ukCol), zap.String("data_type", colDef.DataType))
				return nil, errors.New(errMsg)
			}
		}
		theUK := mcp.NewUniqueKey(-1, ukValues)
		theMCP.AddUK(theUK)
	}
	return theMCP, nil
}
