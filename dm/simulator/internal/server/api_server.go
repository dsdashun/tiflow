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

package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/errors"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/openapi"
	"go.uber.org/zap"
)

// APIServer implements all kinds of APIs.
type APIServer struct {
	parentServer *Server
}

// NewAPIServer creates a new API server.
func NewAPIServer(parentServer *Server) *APIServer {
	return &APIServer{
		parentServer: parentServer,
	}
}

// ApplyDDL implements the apply DDL API.
func (s *APIServer) ApplyDDL(c *gin.Context, sourceName string, tableID string) {
	var gerr error
	response := &openapi.BasicResponse{}
	defer func() {
		if gerr != nil {
			response.ReturnCode = 1
			errMsg := gerr.Error()
			response.ErrorMsg = &errMsg
			c.IndentedJSON(http.StatusInternalServerError, response)
		} else {
			c.IndentedJSON(http.StatusOK, response)
		}
	}()
	simu := s.parentServer.GetDBSimulator(sourceName)
	if simu == nil {
		errMsg := "cannot find the data source"
		gerr = errors.New(errMsg)
		plog.L().Error(errMsg, zap.String("data source id", sourceName))
		return
	}
	tblConfig := simu.GetTableConfig(tableID)
	if tblConfig == nil {
		errMsg := "cannot find the table config"
		gerr = errors.New(errMsg)
		plog.L().Error(errMsg, zap.String("table id", tableID))
		return
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER TABLE `%s`.`%s` ", tblConfig.DatabaseName, tblConfig.TableName))
	if _, err := io.Copy(&sb, c.Request.Body); err != nil {
		gerr = errors.Annotate(err, "copy request body into a string builder error")
		return
	}
	db := simu.GetDB()
	if db == nil {
		errMsg := "cannot get the DB handle"
		gerr = errors.New(errMsg)
		plog.L().Error(errMsg, zap.String("data source id", sourceName))
		return
	}
	simuCtx := simu.GetContext()
	if simuCtx == nil {
		errMsg := "cannot get the context of the simulator"
		gerr = errors.New(errMsg)
		plog.L().Error(errMsg, zap.String("data source id", sourceName))
		return
	}
	if _, err := db.ExecContext(simuCtx, sb.String()); err != nil {
		errMsg := "execute the SQL error"
		gerr = errors.Annotate(err, errMsg)
		plog.L().Error(errMsg, zap.String("data source id", sourceName), zap.String("sql", sb.String()))
		return
	}
}
