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
	"fmt"
	"strings"
	"sync"
)

type UniqueKey struct {
	sync.RWMutex
	OPLock sync.Mutex
	RowID  int
	Value  map[string]interface{}
}

func (uk *UniqueKey) Clone() *UniqueKey {
	result := &UniqueKey{
		RowID: uk.RowID,
		Value: map[string]interface{}{},
	}
	uk.RLock()
	defer uk.RUnlock()
	for k, v := range uk.Value {
		result.Value[k] = v
	}
	return result
}

func (uk *UniqueKey) String() string {
	uk.RLock()
	defer uk.RUnlock()
	var b strings.Builder
	fmt.Fprintf(&b, "%p: { RowID: %d, ", uk, uk.RowID)
	fmt.Fprintf(&b, "Keys: ( ")
	for k, v := range uk.Value {
		fmt.Fprintf(&b, "%s = %v; ", k, v)
	}
	fmt.Fprintf(&b, ") }")
	return b.String()
}

func (uk *UniqueKey) IsValueEqual(otherUK *UniqueKey) bool {
	uk.RLock()
	defer uk.RUnlock()
	if otherUK == nil {
		return false
	}
	if len(uk.Value) != len(otherUK.Value) {
		return false
	}
	for k, v := range uk.Value {
		otherV, ok := otherUK.Value[k]
		if !ok {
			return false
		}
		if v != otherV {
			return false
		}
	}
	return true
}

type UniqueKeyIterator interface {
	NextUK() *UniqueKey
}
