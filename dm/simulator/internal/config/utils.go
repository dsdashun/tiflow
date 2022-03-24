package config

import (
	"reflect"
	"sort"
)

func CloneSortedColumnDefinitions(colDefs []*ColumnDefinition) []*ColumnDefinition {
	var sortedColDefs []*ColumnDefinition
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
