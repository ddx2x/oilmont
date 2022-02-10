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

func (i *systemServer) ListRegion(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")

	results, err := i.region.List(name, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateRegion(g *gin.Context) {
	request := &system.Region{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.region.Create(request)
	if err != nil {
		i.RecordEvent(common.REGION, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.REGION, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateRegion(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")

	request := &system.Region{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := i.region.Update(namespace, name, request)
	if err != nil {
		i.RecordEvent(common.REGION, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.REGION, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) DeleteRegion(g *gin.Context) {
	name := g.Param("name")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.region.Delete(name)
	if err != nil {
		i.RecordEvent(common.REGION, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.REGION, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
