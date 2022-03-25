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

type APIServer struct {
	parentServer *Server
}

func NewAPIServer(parentServer *Server) *APIServer {
	return &APIServer{
		parentServer: parentServer,
	}
}

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
	io.Copy(&sb, c.Request.Body)
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
