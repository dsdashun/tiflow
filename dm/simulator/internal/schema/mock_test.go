package schema

import (
	"context"
	"reflect"
	"testing"

	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/stretchr/testify/suite"
)

type testMockSchemaGenSuite struct {
	suite.Suite
}

func (s *testMockSchemaGenSuite) TestGetColDefsBasic() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sg := NewMockSchemaGetter()
	cfg := config.NewTemplateTableConfigForTest()
	sg.SetFromTableConfig(cfg)
	colDefs, err := sg.GetColumnDefinitions(ctx, cfg.DatabaseName, cfg.TableName)
	s.Require().Nil(err)
	s.True(config.AreColDefinitionsEqual(colDefs, cfg.Columns))
	_, err = sg.GetColumnDefinitions(ctx, "NOT_EXIST", "NOT_EXIST")
	s.NotNil(err)
	newColDefs := []*config.ColumnDefinition{
		cfg.Columns[3],
		cfg.Columns[2],
		cfg.Columns[0],
		&config.ColumnDefinition{
			ColumnName: "newdol",
			DataType:   "int",
		},
	}
	s.False(config.AreColDefinitionsEqual(newColDefs, colDefs))
	func() {
		curDataType := cfg.Columns[1].DataType
		cfg.Columns[1].DataType = "int"
		defer func() {
			cfg.Columns[1].DataType = curDataType
		}()
		s.False(config.AreColDefinitionsEqual(cfg.Columns, colDefs))
	}()
	func() {
		curLen := len(cfg.Columns)
		cfg.Columns = append(cfg.Columns, &config.ColumnDefinition{
			ColumnName: "newdol",
			DataType:   "int",
		})
		defer func() {
			cfg.Columns = cfg.Columns[:curLen]
		}()
		s.False(config.AreColDefinitionsEqual(cfg.Columns, colDefs))
	}()
}

func (s *testMockSchemaGenSuite) TestGetUKCols() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sg := NewMockSchemaGetter()
	cfg := config.NewTemplateTableConfigForTest()
	sg.SetFromTableConfig(cfg)
	ukCols, err := sg.GetUniqueKeyColumns(ctx, cfg.DatabaseName, cfg.TableName)
	s.Require().Nil(err)
	s.True(reflect.DeepEqual(ukCols, cfg.UniqueKeyColumnNames))
	_, err = sg.GetUniqueKeyColumns(ctx, "NOT_EXIST", "NOT_EXIST")
	s.NotNil(err)
	func() {
		curColName := cfg.UniqueKeyColumnNames[0]
		cfg.UniqueKeyColumnNames[0] = "team_id"
		defer func() {
			cfg.UniqueKeyColumnNames[0] = curColName
		}()
		s.False(reflect.DeepEqual(ukCols, cfg.UniqueKeyColumnNames))
	}()
}

func TestMockSchemaGenSuite(t *testing.T) {
	suite.Run(t, &testMockSchemaGenSuite{})
}
