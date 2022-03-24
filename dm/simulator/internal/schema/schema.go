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

// Package schema defines logic on manipulating schema of simulating tables.
package schema

import (
	"context"

	"github.com/pingcap/tiflow/dm/simulator/internal/config"
)

// SchemaGetter defines the operations on getting the schema of a table.
type SchemaGetter interface {
	// GetColumnDefinitions gets the column definitions of a table.
	GetColumnDefinitions(ctx context.Context, dbName string, tableName string) ([]*config.ColumnDefinition, error)
	// GetUniqueKeyColumns gets the columns of a unique key in a table.
	GetUniqueKeyColumns(ctx context.Context, dbName string, tableName string) ([]string, error)
}
