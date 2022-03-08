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

package core

import (
	"context"
	"database/sql"
	"sync/atomic"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/pingcap/errors"
	"go.uber.org/zap"

	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/parser"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type workloadSimulatorImpl struct {
	steps            []DMLWorkloadStep
	totalExecutedTrx uint64
	tblConfigs       map[string]*config.TableConfig
}

func NewWorkloadSimulatorImpl(
	tblConfigs map[string]*config.TableConfig,
	workloadCode string,
) (*workloadSimulatorImpl, error) {
	var err error
	input := antlr.NewInputStream(workloadCode)
	lexer := parser.NewWorkloadLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewWorkloadParser(stream)
	el := NewParseStepsErrorListener(
		antlr.NewDiagnosticErrorListener(true),
	)
	p.AddErrorListener(el)
	p.BuildParseTrees = true
	tree := p.WorkloadSteps()
	sl := NewParseStepsListener(tblConfigs)
	antlr.ParseTreeWalkerDefault.Walk(sl, tree)
	err = el.Err()
	if err != nil {
		return nil, errors.Annotate(err, "parse workload DSL code error")
	}

	involvedTblConfigs := make(map[string]*config.TableConfig)
	for _, step := range sl.totalSteps {
		tblName := step.GetTableName()
		if _, ok := involvedTblConfigs[tblName]; ok {
			continue
		}
		if _, ok := tblConfigs[tblName]; !ok {
			err = ErrTableConfigNotFound
			plog.L().Error(err.Error(), zap.String("table_name", tblName))
			return nil, err
		}
		involvedTblConfigs[tblName] = tblConfigs[tblName]
	}

	return &workloadSimulatorImpl{
		steps:      sl.totalSteps,
		tblConfigs: involvedTblConfigs,
	}, nil
}

func (s *workloadSimulatorImpl) SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*sqlgen.ModificationCandidatePool) error {
	var err error
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Annotate(err, "begin trx error when simulating a trx")
	}

	sctx := &DMLWorkloadStepContext{
		tx:      tx,
		ctx:     ctx,
		rowRefs: make(map[string]*sqlgen.UniqueKey),
	}
	for _, step := range s.steps {
		tblName := step.GetTableName()
		mcp := mcpMap[tblName]
		sctx.mcp = mcp
		err = step.Execute(sctx)
		if err != nil {
			tx.Rollback()
			return errors.Annotate(err, "execute the workload step error")
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.Annotate(err, "trx COMMIT error when simulating a trx")
	}
	atomic.AddUint64(&s.totalExecutedTrx, 1)
	return nil
}

func (s *workloadSimulatorImpl) GetInvolvedTables() []string {
	involvedTbls := []string{}
	for tblName := range s.tblConfigs {
		involvedTbls = append(involvedTbls, tblName)
	}
	return involvedTbls
}
