package sqlgen

import (
	"strings"

	"github.com/chaos-mesh/go-sqlsmith/util"
	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
	"github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tidb/parser/opcode"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"go.uber.org/zap"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
)

type sqlGeneratorImpl struct {
	tableConfig *config.TableConfig
	columnMap   map[string]*config.ColumnDefinition
	ukMap       map[string]struct{}
}

func NewSQLGeneratorImpl(tableConfig *config.TableConfig) *sqlGeneratorImpl {
	colDefMap := make(map[string]*config.ColumnDefinition)
	for _, colDef := range tableConfig.Columns {
		colDefMap[colDef.ColumnName] = colDef
	}
	ukMap := make(map[string]struct{})
	for _, ukColName := range tableConfig.UniqueKeyColumnNames {
		if _, ok := colDefMap[ukColName]; ok {
			ukMap[ukColName] = struct{}{}
		}
	}
	return &sqlGeneratorImpl{
		tableConfig: tableConfig,
		columnMap:   colDefMap,
		ukMap:       ukMap,
	}
}

// outputString parser ast node to SQL string
func outputString(node ast.Node) (string, error) {
	var sb strings.Builder
	err := node.Restore(format.NewRestoreCtx(format.DefaultRestoreFlags, &sb))
	if err != nil {
		return "", errors.Annotate(err, "restore AST into SQL string error")
	}
	return sb.String(), nil
}

func (g *sqlGeneratorImpl) GenTruncateTable() (string, error) {
	truncateTree := &ast.TruncateTableStmt{
		Table: &ast.TableName{
			Schema: model.NewCIStr(g.tableConfig.DatabaseName),
			Name:   model.NewCIStr(g.tableConfig.TableName),
		},
	}
	return outputString(truncateTree)
}

func (g *sqlGeneratorImpl) generateWhereClause(theUK map[string]interface{}) (ast.ExprNode, error) {
	compareExprs := make([]*ast.BinaryOperationExpr, 0)
	//iterate the existing UKs, to make sure all the uk columns has values
	for ukColName := range g.ukMap {
		val, ok := theUK[ukColName]
		if !ok {
			log.L().Error(ErrUKColValueNotProvided.Error(), zap.String("column_name", ukColName))
			return nil, errors.Trace(ErrUKColValueNotProvided)
		}
		compareExprs = append(compareExprs, &ast.BinaryOperationExpr{
			Op: opcode.EQ,
			L: &ast.ColumnNameExpr{
				Name: &ast.ColumnName{
					Name: model.NewCIStr(ukColName),
				},
			},
			R: ast.NewValueExpr(val, "", ""),
		})
	}
	return generateCompoundBinaryOpExpr(compareExprs), nil
}

func generateCompoundBinaryOpExpr(compExprs []*ast.BinaryOperationExpr) ast.ExprNode {
	switch len(compExprs) {
	case 0:
		return nil
	case 1:
		return compExprs[0]
	default:
		return &ast.BinaryOperationExpr{
			Op: opcode.LogicAnd,
			L:  compExprs[0],
			R:  generateCompoundBinaryOpExpr(compExprs[1:]),
		}
	}
}

func (g *sqlGeneratorImpl) GenUpdateRow(theUK *UniqueKey) (string, error) {
	if theUK == nil {
		return "", errors.Trace(ErrMissingUKValue)
	}
	assignments := make([]*ast.Assignment, 0)
	for _, colInfo := range g.columnMap {
		if _, ok := g.ukMap[colInfo.ColumnName]; ok {
			//this is a UK column, skip from modifying it
			continue
		}
		assignments = append(assignments, &ast.Assignment{
			Column: &ast.ColumnName{
				Name: model.NewCIStr(colInfo.ColumnName),
			},
			Expr: ast.NewValueExpr(util.GenerateDataItem(colInfo.DataType), "", ""),
		})
	}
	whereClause, err := g.generateWhereClause(theUK.Value)
	if err != nil {
		return "", errors.Annotate(err, "generate where clause error")
	}
	updateTree := &ast.UpdateStmt{
		List: assignments,
		TableRefs: &ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.TableName{
					Schema: model.NewCIStr(g.tableConfig.DatabaseName),
					Name:   model.NewCIStr(g.tableConfig.TableName),
				},
			},
		},
		Where: whereClause,
	}
	return outputString(updateTree)
}

func (g *sqlGeneratorImpl) GenInsertRow() (string, *UniqueKey, error) {
	ukValues := make(map[string]interface{})
	columnNames := []*ast.ColumnName{}
	values := []ast.ExprNode{}
	for _, col := range g.columnMap {
		columnNames = append(columnNames, &ast.ColumnName{
			Name: model.NewCIStr(col.ColumnName),
		})
		newValue := util.GenerateDataItem(col.DataType)
		values = append(values, ast.NewValueExpr(newValue, "", ""))
		if _, ok := g.ukMap[col.ColumnName]; ok {
			//add UK value
			ukValues[col.ColumnName] = newValue
		}
	}
	insertTree := &ast.InsertStmt{
		Table: &ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.TableName{
					Schema: model.NewCIStr(g.tableConfig.DatabaseName),
					Name:   model.NewCIStr(g.tableConfig.TableName),
				},
			},
		},
		Lists:   [][]ast.ExprNode{values},
		Columns: columnNames,
	}
	sql, err := outputString(insertTree)
	if err != nil {
		return "", nil, errors.Annotate(err, "output INSERT AST into SQL string error")
	} else {
		return sql,
			&UniqueKey{
				RowID: -1,
				Value: ukValues,
			},
			nil
	}
}

func (g *sqlGeneratorImpl) GenDeleteRow(theUK *UniqueKey) (string, error) {
	if theUK == nil {
		return "", errors.Trace(ErrMissingUKValue)
	}
	whereClause, err := g.generateWhereClause(theUK.Value)
	if err != nil {
		return "", errors.Annotate(err, "generate where clause error")
	}
	updateTree := &ast.DeleteStmt{
		TableRefs: &ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.TableName{
					Schema: model.NewCIStr(g.tableConfig.DatabaseName),
					Name:   model.NewCIStr(g.tableConfig.TableName),
				},
			},
		},
		Where: whereClause,
	}
	return outputString(updateTree)
}

func (g *sqlGeneratorImpl) GenLoadUniqueKeySQL() (string, []*config.ColumnDefinition, error) {
	selectFields := make([]*ast.SelectField, 0)
	cols := make([]*config.ColumnDefinition, 0)
	for ukColName := range g.ukMap {
		selectFields = append(selectFields, &ast.SelectField{
			Expr: &ast.ColumnNameExpr{
				Name: &ast.ColumnName{
					Name: model.NewCIStr(ukColName),
				},
			},
		})
		cols = append(cols, g.columnMap[ukColName])
	}
	selectTree := &ast.SelectStmt{
		SelectStmtOpts: &ast.SelectStmtOpts{
			SQLCache: true,
		},
		Fields: &ast.FieldList{
			Fields: selectFields,
		},
		From: &ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.TableName{
					Schema: model.NewCIStr(g.tableConfig.DatabaseName),
					Name:   model.NewCIStr(g.tableConfig.TableName),
				},
			},
		},
	}
	sql, err := outputString(selectTree)
	if err != nil {
		return "", nil, errors.Annotate(err, "output SELECT AST into SQL string error")
	}
	return sql, cols, nil
}
