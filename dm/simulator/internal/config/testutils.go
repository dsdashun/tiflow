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

// NewTemplateTableConfigForTest generates a `TableConfig` for test.
// So that this logic is reused by all kinds of unit tests.
func NewTemplateTableConfigForTest() *TableConfig {
	return &TableConfig{
		TableID:      "members",
		DatabaseName: "games",
		TableName:    "members",
		Columns: []*ColumnDefinition{
			{
				ColumnName: "id",
				DataType:   "int",
			},
			{
				ColumnName: "name",
				DataType:   "varchar",
			},
			{
				ColumnName: "age",
				DataType:   "int",
			},
			{
				ColumnName: "team_id",
				DataType:   "int",
			},
		},
		UniqueKeyColumnNames: []string{"id"},
	}
}
