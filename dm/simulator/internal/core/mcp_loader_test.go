package core

import (
	"context"
	"database/sql/driver"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/stretchr/testify/suite"
)

type testMCPLoaderSuite struct {
	suite.Suite
}

func (s *testMCPLoaderSuite) SetupSuite() {
	s.Require().Nil(log.InitLogger(&log.Config{}))
}

func (s *testMCPLoaderSuite) TestMockMCPLoader() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewTemplateTableConfigForTest()
	ukColMap := make(map[string]struct{})
	for _, ukCol := range cfg.UniqueKeyColumnNames {
		ukColMap[ukCol] = struct{}{}
	}
	recordCount := 128
	ml := NewMockMCPLoader(recordCount)
	theMCP, err := ml.LoadMCP(ctx, cfg)
	s.Require().Nil(err)
	s.Equal(recordCount, theMCP.Len())
	for i := 0; i < theMCP.Len(); i++ {
		theUK := theMCP.GetUK(i)
		s.Require().NotEqual(-1, theUK.GetRowID())
		ukVal := theUK.GetValue()
		s.Require().Equal(len(ukVal), len(ukColMap))
		for ukCol := range ukVal {
			_, ok := ukColMap[ukCol]
			s.Require().True(ok)
		}
	}
}

func (s *testMCPLoaderSuite) TestLoadMCP() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewTemplateTableConfigForTest()
	ukColMap := make(map[string]struct{})
	for _, ukCol := range cfg.UniqueKeyColumnNames {
		ukColMap[ukCol] = struct{}{}
	}

	db, mock, err := sqlmock.New()
	s.Require().Nil(err)
	recordCount := 128
	mockML := NewMockMCPLoader(recordCount)
	mockMCP, err := mockML.LoadMCP(ctx, cfg)

	ml := NewMCPLoaderImpl(db)
	expectRows := sqlmock.NewRows(cfg.UniqueKeyColumnNames)
	for i := 0; i < recordCount; i++ {
		theUK := mockMCP.GetUK(i)
		s.Require().NotNil(theUK)
		ukValMap := theUK.GetValue()
		var values []driver.Value
		for _, ukCol := range cfg.UniqueKeyColumnNames {
			val, ok := ukValMap[ukCol]
			s.Require().True(ok)
			values = append(values, driver.Value(val))
		}
		expectRows.AddRow(values...)
	}
	mock.ExpectQuery("^SELECT").WillReturnRows(expectRows)
	theMCP, err := ml.LoadMCP(ctx, cfg)
	s.Require().Nil(err)
	s.Equal(recordCount, theMCP.Len())
	for i := 0; i < theMCP.Len(); i++ {
		theUK := theMCP.GetUK(i)
		s.Require().NotEqual(-1, theUK.GetRowID())
		ukVal := theUK.GetValue()
		s.Require().Equal(len(ukVal), len(ukColMap))
		for ukCol := range ukVal {
			_, ok := ukColMap[ukCol]
			s.Require().True(ok)
		}
	}
}

func TestMCPLoaderSuite(t *testing.T) {
	suite.Run(t, &testMCPLoaderSuite{})
}
