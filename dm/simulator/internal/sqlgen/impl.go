package sqlgen

import (
	"strings"

	"github.com/chaos-mesh/go-sqlsmith/types"
	"github.com/chaos-mesh/go-sqlsmith/util"
	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
	"github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tidb/parser/opcode"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

type sqlGeneratorImpl struct {
	tableInfo *types.Table
	ukColumns map[string]*types.Column
}

func NewSQLGeneratorImpl(tableInfo *types.Table, ukColumns map[string]*types.Column) *sqlGeneratorImpl {
	return &sqlGeneratorImpl{
		tableInfo: tableInfo,
		ukColumns: ukColumns,
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
			Schema: model.NewCIStr(g.tableInfo.DB),
			Name:   model.NewCIStr(g.tableInfo.Table),
		},
	}
	return outputString(truncateTree)
}

func (g *sqlGeneratorImpl) generateWhereClause(theUK map[string]interface{}) ast.ExprNode {
	compareExprs := make([]*ast.BinaryOperationExpr, 0)
	for colName, val := range theUK {
		keyColumn := g.tableInfo.Columns[colName]
		compareExprs = append(compareExprs, &ast.BinaryOperationExpr{
			Op: opcode.EQ,
			L: &ast.ColumnNameExpr{
				Name: &ast.ColumnName{
					Name: model.NewCIStr(keyColumn.Column),
				},
			},
			R: ast.NewValueExpr(val, "", ""),
		})
	}
	return generateCompoundBinaryOpExpr(compareExprs)
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
	ukValues := make(map[string]interface{})
	// check the input UK's correctness.  Currently only check the column name
	for _, colDef := range g.ukColumns {
		ukVal, ok := theUK.Value[colDef.Column]
		if !ok {
			return "", errors.Trace(ErrUKColumnsMismatch)
		}
		ukValues[colDef.Column] = ukVal
	}
	assignments := make([]*ast.Assignment, 0)
	for _, colInfo := range g.tableInfo.Columns {
		if _, ok := g.ukColumns[colInfo.Column]; ok {
			continue
		}
		assignments = append(assignments, &ast.Assignment{
			Column: &ast.ColumnName{
				Name: model.NewCIStr(colInfo.Column),
			},
			Expr: ast.NewValueExpr(util.GenerateDataItem(colInfo.DataType), "", ""),
		})
	}
	updateTree := &ast.UpdateStmt{
		List: assignments,
		TableRefs: &ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.TableName{
					Schema: model.NewCIStr(g.tableInfo.DB),
					Name:   model.NewCIStr(g.tableInfo.Table),
				},
			},
		},
		Where: g.generateWhereClause(ukValues),
	}
	return outputString(updateTree)
}

func (g *sqlGeneratorImpl) GenInsertRow() (string, *UniqueKey, error) {
	ukValues := make(map[string]interface{})
	columnNames := []*ast.ColumnName{}
	values := []ast.ExprNode{}
	for _, col := range g.tableInfo.Columns {
		columnNames = append(columnNames, &ast.ColumnName{
			Name: model.NewCIStr(col.Column),
		})
		newValue := util.GenerateDataItem(col.DataType)
		values = append(values, ast.NewValueExpr(newValue, "", ""))
		if _, ok := g.ukColumns[col.Column]; ok {
			ukValues[col.Column] = newValue
		}
	}
	insertTree := &ast.InsertStmt{
		Table: &ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.TableName{
					Schema: model.NewCIStr(g.tableInfo.DB),
					Name:   model.NewCIStr(g.tableInfo.Table),
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
	ukValues := make(map[string]interface{})
	// check the input UK's correctness.  Currently only check the column name
	for _, colDef := range g.ukColumns {
		ukVal, ok := theUK.Value[colDef.Column]
		if !ok {
			return "", errors.Trace(ErrUKColumnsMismatch)
		}
		ukValues[colDef.Column] = ukVal
	}
	updateTree := &ast.DeleteStmt{
		TableRefs: &ast.TableRefsClause{
			TableRefs: &ast.Join{
				Left: &ast.TableName{
					Schema: model.NewCIStr(g.tableInfo.DB),
					Name:   model.NewCIStr(g.tableInfo.Table),
				},
			},
		},
		Where: g.generateWhereClause(ukValues),
	}
	return outputString(updateTree)
}

func (g *sqlGeneratorImpl) GenLoadUniqueKeySQL() (string, []*types.Column, error) {
	selectFields := make([]*ast.SelectField, 0)
	cols := make([]*types.Column, 0)
	for _, ukCol := range g.ukColumns {
		selectFields = append(selectFields, &ast.SelectField{
			Expr: &ast.ColumnNameExpr{
				Name: &ast.ColumnName{
					Name: model.NewCIStr(ukCol.Column),
				},
			},
		})
		cols = append(cols, ukCol)
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
					Schema: model.NewCIStr(g.tableInfo.DB),
					Name:   model.NewCIStr(g.tableInfo.Table),
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
