package system

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
)

func (i *systemServer) ListResource(g *gin.Context) {
	results, err := i.resourceService.List("")
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *systemServer) GetResource(g *gin.Context) {
	name := g.Param("name")

	results, err := i.resourceService.List(name)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateResource(g *gin.Context) {
	request := &system.Resource{}

	err := g.ShouldBindJSON(request)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.resourceService.Create(request)
	if err != nil {
		i.RecordEvent(common.RESOURCE, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.RESOURCE, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateResource(g *gin.Context) {
	name := g.Param("name")
	paths := api.SplitPath(g.Query("path"))

	request := &system.Resource{}
	err := g.ShouldBindJSON(request)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	res, isUpdate, err := i.resourceService.Update(name, request, paths...)
	_ = isUpdate
	if err != nil {
		i.RecordEvent(common.RESOURCE, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.RESOURCE, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) DeleteResource(g *gin.Context) {
	name := g.Param("name")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.resourceService.Delete(name)
	if err != nil {
		i.RecordEvent(common.RESOURCE, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.RESOURCE, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
