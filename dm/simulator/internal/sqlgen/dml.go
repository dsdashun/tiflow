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

package sqlgen

import (
	"github.com/chaos-mesh/go-sqlsmith/types"
	"github.com/chaos-mesh/go-sqlsmith/util"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tidb/parser/opcode"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

type dmlSQLGenerator struct {
	tableInfo *types.Table
	ukColumns map[string]*types.Column
}

func NewDMLSQLGenerator() *dmlSQLGenerator {
	tableInfo := &types.Table{
		DB:    "games",
		Table: "members",
		Type:  "BASE TABLE",
		Columns: map[string]*types.Column{
			"id": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "id",
				DataType: "int",
				DataLen:  11,
			},
			"name": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "name",
				DataType: "varchar",
				DataLen:  255,
			},
			"age": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "age",
				DataType: "int",
				DataLen:  11,
			},
			"team_id": &types.Column{
				DB:       "games",
				Table:    "members",
				Column:   "team_id",
				DataType: "int",
				DataLen:  11,
			},
		},
	}
	ukColumns := map[string]*types.Column{
		"id": tableInfo.Columns["id"],
	}
	return &dmlSQLGenerator{
		tableInfo: tableInfo,
		ukColumns: ukColumns,
	}
}

func (g *dmlSQLGenerator) generateWhereClause(theUK map[string]interface{}) ast.ExprNode {
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

func (g *dmlSQLGenerator) GenUpdateRow(theUK *UniqueKey) (string, error) {
	ukValues := make(map[string]interface{})
	// check the input UK's correctness.  Currently only check the column name
	for _, colDef := range g.ukColumns {
		ukVal, ok := theUK.Value[colDef.Column]
		if !ok {
			return "", ErrUKColumnsMismatch
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
	return util.BufferOut(updateTree)
}

func (g *dmlSQLGenerator) GenInsertRow() (string, *UniqueKey, error) {
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
	sql, err := util.BufferOut(insertTree)
	if err != nil {
		return "", nil, err
	} else {
		return sql,
			&UniqueKey{
				RowID: -1,
				Value: ukValues,
			},
			nil
	}
}

func (g *dmlSQLGenerator) GenDeleteRow(theUK *UniqueKey) (string, error) {
	ukValues := make(map[string]interface{})
	// check the input UK's correctness.  Currently only check the column name
	for _, colDef := range g.ukColumns {
		ukVal, ok := theUK.Value[colDef.Column]
		if !ok {
			return "", ErrUKColumnsMismatch
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
	return util.BufferOut(updateTree)
}
