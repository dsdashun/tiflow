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

// Package sqlgen is the logic for generating different kinds of SQL statements.
package sqlgen

import "github.com/pingcap/tiflow/dm/simulator/internal/config"

// DMLType is the type describing different kinds of DML operations.
type DMLType int

const (
	// DMLTypeUNKNOWN is for the unknown DML.
	DMLTypeUNKNOWN DMLType = iota
	// DMLTypeINSERT is for the INSERT DML.
	DMLTypeINSERT
	// DMLTypeUPDATE is for the UPDATE DML.
	DMLTypeUPDATE
	// DMLTypeDELETE is for the DELETE DML.
	DMLTypeDELETE
)

// SQLGenerator contains all the operations for generating SQLs.
type SQLGenerator interface {
	// GenTruncateTable generates a TRUNCATE TABLE SQL.
	GenTruncateTable() (string, error)
	// GenLoadUniqueKeySQL generates a SELECT SQL fetching all the uniques of a table.
	// The column definitions of the returned data is also provided,
	// so that the values can be stored to different variables.
	GenLoadUniqueKeySQL() (string, []*config.ColumnDefinition, error)
	// GenInsertRow generates an INSERT SQL.
	// The new row's unique key is also provided,
	// so that it can be further added into an MCP.
	GenInsertRow() (string, *UniqueKey, error)
	// GenUpdateRow generates an UPDATE SQL for the given unique key.
	GenUpdateRow(*UniqueKey) (string, error)
	// GenDeleteRow generates a DELETE SQL for the given unique key.
	GenDeleteRow(*UniqueKey) (string, error)
}
