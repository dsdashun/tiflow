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

package config

import (
	"reflect"
	"sort"
)

// CloneSortedColumnDefinitions clones column definitions to a new list with column name sorted.
func CloneSortedColumnDefinitions(colDefs []*ColumnDefinition) []*ColumnDefinition {
	sortedColDefs := make([]*ColumnDefinition, 0)
	for _, colDef := range colDefs {
		sortedColDefs = append(sortedColDefs, &ColumnDefinition{
			ColumnName: colDef.ColumnName,
			DataType:   colDef.DataType,
		})
	}
	sort.Slice(sortedColDefs, func(i, j int) bool {
		return sortedColDefs[i].ColumnName < sortedColDefs[j].ColumnName
	})
	return sortedColDefs
}

// AreColDefinitionsEqual checks whether two column definitions are actually equal.
func AreColDefinitionsEqual(colDefs01 []*ColumnDefinition, colDefs02 []*ColumnDefinition) bool {
	if colDefs01 == nil && colDefs02 == nil {
		return true
	}
	if colDefs01 == nil || colDefs02 == nil {
		return false
	}
	if len(colDefs01) != len(colDefs02) {
		return false
	}
	sortedColDefs01 := CloneSortedColumnDefinitions(colDefs01)
	sortedColDefs02 := CloneSortedColumnDefinitions(colDefs02)
	return reflect.DeepEqual(sortedColDefs01, sortedColDefs02)
}

// GenerateColumnDefinitionsMap converts a series of column definitions into a column definition map with column name as the key.
func GenerateColumnDefinitionsMap(colDef []*ColumnDefinition) map[string]*ColumnDefinition {
	colDefMap := make(map[string]*ColumnDefinition)
	for _, colDef := range colDef {
		colDefMap[colDef.ColumnName] = &ColumnDefinition{
			ColumnName: colDef.ColumnName,
			DataType:   colDef.DataType,
		}
	}
	return colDefMap
}
