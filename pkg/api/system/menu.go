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

func (i *systemServer) ListMenu(g *gin.Context) {
	results, err := i.menu.List()
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateMenu(g *gin.Context) {
	request := &system.Menu{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.menu.Create(request)
	if err != nil {
		i.RecordEvent(common.Menu, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.Menu, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateMenu(g *gin.Context) {
	name := g.Param("name")
	path := g.Query("path")
	request := &system.Menu{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	new, isUpdate, err := i.menu.Update(name, request, path)
	if err != nil {
		i.RecordEvent(common.Menu, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}
	_ = isUpdate
	i.RecordEvent(common.Menu, core.MODIFIED, reqUser, new, event.CloudEventSuccess)
	g.JSON(http.StatusOK, new)
}

func (i *systemServer) DeleteMenu(g *gin.Context) {
	name := g.Param("name")
	res, err := i.menu.Delete(name)
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	if err != nil {
		i.RecordEvent(common.Menu, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.Menu, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
