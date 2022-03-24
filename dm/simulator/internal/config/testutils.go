package config

func NewTemplateTableConfigForTest() *TableConfig {
	return &TableConfig{
		TableID:      "members",
		DatabaseName: "games",
		TableName:    "members",
		Columns: []*ColumnDefinition{
			&ColumnDefinition{
				ColumnName: "id",
				DataType:   "int",
			},
			&ColumnDefinition{
				ColumnName: "name",
				DataType:   "varchar",
			},
			&ColumnDefinition{
				ColumnName: "age",
				DataType:   "int",
			},
			&ColumnDefinition{
				ColumnName: "team_id",
				DataType:   "int",
			},
		},
		UniqueKeyColumnNames: []string{"id"},
	}
}
