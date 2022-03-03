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

import "github.com/pingcap/tiflow/dm/simulator/internal/config"

type DMLType int

const (
	UNKNOWN_DMLType DMLType = iota
	INSERT_DMLType
	UPDATE_DMLType
	DELETE_DMLType
)

type SQLGenerator interface {
	GenTruncateTable() (string, error)
	GenLoadUniqueKeySQL() (string, []*config.ColumnDefinition, error)
	GenInsertRow() (string, *UniqueKey, error)
	GenUpdateRow(*UniqueKey) (string, error)
	GenDeleteRow(*UniqueKey) (string, error)
}
