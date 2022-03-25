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

// Package config is the configuration definitions used by the simulator.
package config

import (
	"os"
	"reflect"

	"github.com/pingcap/errors"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// CLIConfig is the configuration struct for command-line-interface options.
type CLIConfig struct {
	IsHelp     bool
	ConfigFile string
}

// NewCLIConfig generates a new CLI config for being parsed.
func NewCLIConfig() *CLIConfig {
	theConf := new(CLIConfig)
	flag.StringVarP(&(theConf.ConfigFile), "config-file", "c", "", "config YAML file")
	flag.BoolVarP(&(theConf.IsHelp), "help", "h", false, "print help message")
	return theConf
}

// Config is the core configurations used by the simulator.
type Config struct {
	DataSources []*DataSourceConfig `yaml:"data_sources"`
	Workloads   []*WorkloadConfig   `yaml:"workloads"`
}

// NewConfigFromFile generates a new config object from a configuration file.
// The config file is in YAML format.
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

// DataSourceConfig is the sub config for describing a DB data source.
type DataSourceConfig struct {
	DataSourceID string         `yaml:"id"`
	Host         string         `yaml:"host"`
	Port         int            `yaml:"port"`
	UserName     string         `yaml:"user"`
	Password     string         `yaml:"password"`
	Tables       []*TableConfig `yaml:"tables"`
}

// TableConfig is the sub config for describing a simulating table in the data source.
type TableConfig struct {
	TableID              string `yaml:"id"`
	DatabaseName         string `yaml:"db"`
	TableName            string `yaml:"table"`
	Columns              []*ColumnDefinition
	UniqueKeyColumnNames []string
}

func (cfg *TableConfig) SortedClone() *TableConfig {
	return &TableConfig{
		TableID:              cfg.TableID,
		DatabaseName:         cfg.DatabaseName,
		TableName:            cfg.TableName,
		Columns:              CloneSortedColumnDefinitions(cfg.Columns),
		UniqueKeyColumnNames: append([]string{}, cfg.UniqueKeyColumnNames...),
	}
}

func (cfgA *TableConfig) IsDeepEqual(cfgB *TableConfig) bool {
	if cfgA == nil || cfgB == nil {
		return false
	}
	if cfgA.TableID != cfgB.TableID ||
		cfgA.DatabaseName != cfgB.DatabaseName ||
		cfgA.TableName != cfgB.TableName {
		return false
	}
	if !AreColDefinitionsEqual(cfgA.Columns, cfgB.Columns) {
		return false
	}
	// begin to check the unique key names
	if !reflect.DeepEqual(cfgA.UniqueKeyColumnNames, cfgB.UniqueKeyColumnNames) {
		return false
	}
	return true
}

// ColumnDefinition is the sub config for describing a column in a simulating table.
type ColumnDefinition struct {
	ColumnName string `yaml:"name"`
	DataType   string `yaml:"type"`
}

// WorkloadConfig is the configuration to describe the attributes of a workload.
type WorkloadConfig struct {
	WorkloadCode string `yaml:"dsl_code"`
}
