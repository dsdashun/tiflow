package core

import (
	"context"
	"math/rand"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

type testWorkloadStepSuite struct {
	suite.Suite
	tableConfig *config.TableConfig
	mcp         *sqlgen.ModificationCandidatePool
}

func (s *testWorkloadStepSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
	s.tableConfig = &config.TableConfig{
		DatabaseName: "games",
		TableName:    "members",
		Columns: []*config.ColumnDefinition{
			&config.ColumnDefinition{
				ColumnName: "id",
				DataType:   "int",
				DataLen:    11,
			},
			&config.ColumnDefinition{
				ColumnName: "name",
				DataType:   "varchar",
				DataLen:    255,
			},
			&config.ColumnDefinition{
				ColumnName: "age",
				DataType:   "int",
				DataLen:    11,
			},
			&config.ColumnDefinition{
				ColumnName: "team_id",
				DataType:   "int",
				DataLen:    11,
			},
		},
		UniqueKeyColumnNames: []string{"id"},
	}
	s.mcp = sqlgen.NewModificationCandidatePool()
	for i := 0; i < 100; i++ {
		s.mcp.AddUK(&sqlgen.UniqueKey{
			RowID: -1,
			Value: map[string]interface{}{
				"id": rand.Int(),
			},
		})
	}
}

func (s *testWorkloadStepSuite) TestBasic() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(s.tableConfig)
	theInsertStep := &insertStep{
		sqlGen: sqlGen,
	}
	theUpdateStep := &updateStep{
		sqlGen: sqlGen,
	}
	theDeleteStep := &deleteStep{
		sqlGen: sqlGen,
	}
	theRandomStep := &randomDMLStep{
		sqlGen: sqlGen,
	}

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	assert.Nil(s.T(), err)

	sctx := &DMLWorkloadStepContext{
		tx:      tx,
		ctx:     ctx,
		mcp:     s.mcp,
		rowRefs: make(map[string]*sqlgen.UniqueKey),
	}
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theInsertStep.Execute(sctx)
	assert.Nil(s.T(), err)

	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theUpdateStep.Execute(sctx)
	assert.Nil(s.T(), err)

	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theDeleteStep.Execute(sctx)
	assert.Nil(s.T(), err)

	mock.ExpectExec("^(INSERT|UPDATE|DELETE) (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theRandomStep.Execute(sctx)
	assert.Nil(s.T(), err)

	mock.ExpectCommit()
	err = tx.Commit()
	assert.Nil(s.T(), err)
}

func (s *testWorkloadStepSuite) TestAssignmentReference() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	sqlGen := sqlgen.NewSQLGeneratorImpl(s.tableConfig)

	theInsertStep := &insertStep{
		sqlGen: sqlGen,
	}

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	assert.Nil(s.T(), err)

	sctx := &DMLWorkloadStepContext{
		tx:      tx,
		ctx:     ctx,
		mcp:     s.mcp,
		rowRefs: make(map[string]*sqlgen.UniqueKey),
	}
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theInsertStep.Execute(sctx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(sctx.rowRefs), "there should be no assigned rows")

	assignedRowID := "@abc01"
	theInsertStep.assignedRowID = assignedRowID

	mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theInsertStep.Execute(sctx)
	assert.Nil(s.T(), err)
	assignedUK, ok := sctx.rowRefs[assignedRowID]
	assert.Equalf(s.T(), true, ok, "%s should be assigned", assignedRowID)
	s.T().Logf("%s assigned with the UK: %v\n", assignedRowID, assignedUK)
	assert.NotEqual(s.T(), -1, assignedUK.RowID, "the new UK should have a valid row ID")

	//normal update
	theUpdateStep := &updateStep{
		sqlGen: sqlGen,
	}
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theUpdateStep.Execute(sctx)
	assert.Nil(s.T(), err)

	//update use the assigned UK
	theUpdateStep.inputRowID = assignedRowID
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theUpdateStep.Execute(sctx)
	assert.Nil(s.T(), err)

	//update use an non-existing UK
	theUpdateStep.inputRowID = "@NOT_EXISTING"
	err = theUpdateStep.Execute(sctx)
	assert.NotNil(s.T(), err, "update a non-existing row should have error")
	s.T().Logf("updating a non-existing rowID get the following error: %v\n", err)

	//double reference the same UK
	assignedRowID2 := "@abc02"
	theUpdateStep.inputRowID = assignedRowID
	theUpdateStep.assignmentRowID = assignedRowID2
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theUpdateStep.Execute(sctx)
	assert.Nil(s.T(), err)
	assignedUK2, ok := sctx.rowRefs[assignedRowID2]
	assert.Equalf(s.T(), true, ok, "%s should be assigned", assignedRowID2)
	s.T().Logf("%s assigned with the UK: %v\n", assignedRowID2, assignedUK2)
	assert.Equal(s.T(), assignedUK, assignedUK2, "the two assignment should be the same")
	assert.GreaterOrEqual(s.T(), assignedUK.RefCount, 2, "the assigned UK should have ref count >=2")

	//delete double-refferred row
	theDeleteStep := &deleteStep{
		sqlGen:     sqlGen,
		inputRowID: assignedRowID2,
	}
	err = theDeleteStep.Execute(sctx)
	assert.Nil(s.T(), err)
	assignedUK2, ok = sctx.rowRefs[assignedRowID2]
	assert.Equalf(s.T(), false, ok, "%s should not be assigned", assignedRowID2)
	s.T().Logf("after delete %s: %v\n", assignedRowID2, assignedUK)
	assert.Equal(s.T(), 1, assignedUK.RefCount, "the row ref should decrease to 1")

	// assign another row to the row ID
	theUpdateStep.assignmentRowID = assignedRowID
	theUpdateStep.inputRowID = ""
	mock.ExpectExec("^UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theUpdateStep.Execute(sctx)
	assert.Nil(s.T(), err)
	anotherAssignedUK, ok := sctx.rowRefs[assignedRowID]
	assert.Equalf(s.T(), true, ok, "%s should be assigned", assignedRowID)
	s.T().Logf("%s assigned with the UK: %v\n", assignedRowID, anotherAssignedUK)
	assert.Equal(s.T(), 1, anotherAssignedUK.RefCount, "the new assigned UK should have ref count of 1")
	if assignedUK != anotherAssignedUK {
		assert.Equal(s.T(), 0, assignedUK.RefCount, "the old assigned UK should have ref count of 0")
	}

	//delete non-existing row-ref
	theDeleteStep.inputRowID = "@NOT_EXISTING"
	err = theDeleteStep.Execute(sctx)
	assert.NotNil(s.T(), err, "delete a non-existing row should have error")
	s.T().Logf("deleting a non-existing rowID get the following error: %v\n", err)

	//delete existing row-ref
	theDeleteStep.inputRowID = assignedRowID
	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theDeleteStep.Execute(sctx)
	assert.Nil(s.T(), err)
	_, ok = sctx.rowRefs[assignedRowID]
	assert.Equalf(s.T(), false, ok, "%s should be unassigned", assignedRowID)
	assert.Equal(s.T(), 0, anotherAssignedUK.RefCount, "the prev assigned UK should have ref count of 0")

	//normal deletion
	theDeleteStep.inputRowID = ""
	mock.ExpectExec("^DELETE (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	err = theDeleteStep.Execute(sctx)
	assert.Nil(s.T(), err)

	mock.ExpectCommit()
	err = tx.Commit()
	assert.Nil(s.T(), err)
}

func TestWorkloadStepSuite(t *testing.T) {
	suite.Run(t, &testWorkloadStepSuite{})
}
