package system

import (
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"net/http"
)

func (i *systemServer) ListWorkspace(g *gin.Context) {
	results, err := i.workspace.List("")
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateWorkspace(g *gin.Context) {
	request := &system.Workspace{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	res, err := i.workspace.Create(request)
	if err != nil {
		i.RecordEvent(common.WORKSPACE, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.WORKSPACE, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateWorkspace(g *gin.Context) {
	name := g.Param("name")
	path := g.Query("path")
	request := &system.Workspace{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	res, _, err := i.workspace.Update(name, request, []string{path}...)
	if err != nil {
		i.RecordEvent(common.WORKSPACE, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}
	i.RecordEvent(common.WORKSPACE, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) DeleteWorkspace(g *gin.Context) {
	name := g.Param("name")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.workspace.Delete(name)
	if err != nil {
		i.RecordEvent(common.WORKSPACE, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.WORKSPACE, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
