package config

import (
	"os"

	"github.com/pingcap/errors"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

type CLIConfig struct {
	IsHelp     bool
	ConfigFile string
}

func NewCLIConfig() *CLIConfig {
	theConf := new(CLIConfig)
	flag.StringVarP(&(theConf.ConfigFile), "config-file", "c", "", "config YAML file")
	flag.BoolVarP(&(theConf.IsHelp), "help", "h", false, "print help message")
	return theConf
}

type Config struct {
	DataSources []*DataSourceConfig `yaml:"data_sources"`
	Workloads   []*WorkloadConfig   `yaml:"workloads"`
}

func NewConfigFromFile(configFile string) (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, errors.Annotate(err, "open config file error")
	}
	theConfig := new(Config)
	dec := yaml.NewDecoder(f)
	err = dec.Decode(theConfig)
	if err != nil {
		return nil, errors.Annotate(err, "decode YAML error")
	}
	return theConfig, nil
}

type DataSourceConfig struct {
	Host     string         `yaml:"host"`
	Port     int            `yaml:"port"`
	UserName string         `yaml:"user"`
	Password string         `yaml:"password"`
	Tables   []*TableConfig `yaml:"tables"`
}

type TableConfig struct {
	TableID              string              `yaml:"id"`
	DatabaseName         string              `yaml:"db"`
	TableName            string              `yaml:"table"`
	Columns              []*ColumnDefinition `yaml:"columns"`
	UniqueKeyColumnNames []string            `yaml:"unique_keys"`
}

type ColumnDefinition struct {
	ColumnName string `yaml:"name"`
	DataType   string `yaml:"type"`
	DataLen    int    `yaml:"length"`
}

type WorkloadConfig struct {
	WorkloadCode string `yaml:"dsl_code"`
}
