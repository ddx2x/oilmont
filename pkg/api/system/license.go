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

func (i *systemServer) ListLicense(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")

	results, err := i.license.List(name, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateLicense(g *gin.Context) {
	request := &system.License{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.license.Create(request)
	if err != nil {
		i.RecordEvent(common.LICENSE, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.LICENSE, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateLicense(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")

	request := &system.License{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := i.license.Update(namespace, name, request)
	if err != nil {
		i.RecordEvent(common.LICENSE, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.LICENSE, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) DeleteLicense(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.license.Delete(namespace, name)
	if err != nil {
		i.RecordEvent(common.LICENSE, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.LICENSE, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
