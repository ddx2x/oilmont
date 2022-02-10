package iam

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
)

func (i *iamServer) ListBusinessGroup(g *gin.Context) {
	namespace := g.Param("namespace")
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	results, err := i.BusinessGroupService.List(tenant, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *iamServer) GetBusinessGroup(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")

	results, err := i.BusinessGroupService.List(name, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *iamServer) CreateBusinessGroup(g *gin.Context) {
	request := &iam.BusinessGroup{}

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.BusinessGroupService.Create(request)
	if err != nil {
		i.RecordEvent(common.BUSINESSGROUP, core.ADDED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.AUTHORITYGROUP, core.ADDED, reqUser, request, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *iamServer) UpdateBusinessGroup(g *gin.Context) {
	opTenant := g.Query("tenant")
	name := g.Param("name")
	paths := api.SplitPath(g.Query("path"))

	request := &iam.BusinessGroup{}

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	res, _, err := i.BusinessGroupService.Update(opTenant, "", name, request, paths...)
	if err != nil {
		i.RecordEvent(common.BUSINESSGROUP, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.BUSINESSGROUP, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *iamServer) DeleteBusinessGroup(g *gin.Context) {
	name := g.Param("name")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	res, err := i.BusinessGroupService.Delete(tenant, "", name)
	if err != nil {
		i.RecordEvent(common.BUSINESSGROUP, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.BUSINESSGROUP, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
