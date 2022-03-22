package schema

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap/errors"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
)

// MySQLSchemaGetter implementes the logic on  getting the schema of a MySQL table.
// It implements the `SchemaGetter` interface.
type MySQLSchemaGetter struct {
	db        *sql.DB
	dbName    string
	tableName string
}

// NewMySQLSchemaGetter generats a new MySQLSchemaGetter instance.
func NewMySQLSchemaGetter(db *sql.DB, dbName string, tableName string) *MySQLSchemaGetter {
	return &MySQLSchemaGetter{
		db:        db,
		dbName:    dbName,
		tableName: tableName,
	}
}

// GetDatabaseName gets the database name of the MySQL table
// It impelements the `SchemaGetter` interface.
func (g *MySQLSchemaGetter) GetDatabaseName() string {
	return g.dbName
}

// GetTableName gets the table name of the table
// It impelements the `SchemaGetter` interface.
func (g *MySQLSchemaGetter) GetTableName() string {
	return g.tableName
}

// GetColumnDefinitions gets the column definitions of a MySQL table.
// It impelements the `SchemaGetter` interface.
func (g *MySQLSchemaGetter) GetColumnDefinitions(ctx context.Context) ([]*config.ColumnDefinition, error) {
	rows, err := g.db.QueryContext(ctx, "SELECT COLUMN_NAME, DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME=?", g.dbName, g.tableName)
	if err != nil {
		return nil, errors.Annotate(err, "query DB error")
	}
	defer rows.Close()
	var result []*config.ColumnDefinition
	for rows.Next() {
		var (
			colName  string
			dataType string
		)
		if err := rows.Scan(&colName, &dataType); err != nil {
			return nil, errors.Annotate(err, "scan the a row into values error")
		}
		result = append(result, &config.ColumnDefinition{
			ColumnName: colName,
			DataType:   dataType,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Annotate(err, "fetching rows error")
	}
	return result, nil
}

type uniqueKeyInfo struct {
	FirstKeyCardinality int
	KeyColumnNames      []string
}

// GetUniqueKeyColumns gets the columns of a unique key in a MySQL table.
// It impelements the `SchemaGetter` interface.
func (g *MySQLSchemaGetter) GetUniqueKeyColumns(ctx context.Context) ([]string, error) {
	rows, err := g.db.QueryContext(ctx, fmt.Sprintf("SHOW INDEX FROM %s.%s WHERE Non_unique=0", g.dbName, g.tableName))
	if err != nil {
		return nil, errors.Annotate(err, "query DB error")
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, errors.Annotate(err, "get columns error")
	}
	var valueHolders []interface{}
	var (
		keyName     string
		seqInIdx    int
		idxColName  string
		cardinality int
	)
	for _, colName := range cols {
		switch colName {
		case "Key_name":
			valueHolders = append(valueHolders, &keyName)
		case "Seq_in_index":
			valueHolders = append(valueHolders, &seqInIdx)
		case "Column_name":
			valueHolders = append(valueHolders, &idxColName)
		case "Cardinality":
			valueHolders = append(valueHolders, &cardinality)
		default:
			valueHolders = append(valueHolders, new(sql.RawBytes))
		}
	}
	allUniqueKeys := make(map[string]*uniqueKeyInfo)
	for rows.Next() {
		if err := rows.Scan(valueHolders...); err != nil {
			return nil, errors.Annotate(err, "scan the a row into values error")
		}
		if _, ok := allUniqueKeys[keyName]; !ok {
			allUniqueKeys[keyName] = &uniqueKeyInfo{
				KeyColumnNames: make([]string, 0),
			}
		}
		theUKInfo := allUniqueKeys[keyName]
		if seqInIdx == 1 {
			theUKInfo.FirstKeyCardinality = cardinality
		}
		if len(theUKInfo.KeyColumnNames) < seqInIdx {
			theUKInfo.KeyColumnNames = append(theUKInfo.KeyColumnNames,
				make([]string, seqInIdx-len(theUKInfo.KeyColumnNames))...,
			)
		}
		theUKInfo.KeyColumnNames[seqInIdx-1] = idxColName
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Annotate(err, "fetching rows error")
	}
	var resultUKInfo *uniqueKeyInfo
	for _, ukInfo := range allUniqueKeys {
		if resultUKInfo == nil {
			resultUKInfo = ukInfo
			continue
		}
		// choose the UK with the most cardinality and the least column numbers
		if ukInfo.FirstKeyCardinality > resultUKInfo.FirstKeyCardinality ||
			(ukInfo.FirstKeyCardinality == resultUKInfo.FirstKeyCardinality &&
				len(ukInfo.KeyColumnNames) < len(resultUKInfo.KeyColumnNames)) {
			resultUKInfo = ukInfo
		}
	}
	if resultUKInfo == nil {
		return nil, nil
	}
	return resultUKInfo.KeyColumnNames, nil
}
