package iam

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
)

func (i *iamServer) ListRole(g *gin.Context) {
	namespace := g.Param("namespace")
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	results, err := i.RoleService.List(tenant, namespace, "")
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *iamServer) GetRole(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	results, err := i.RoleService.List(tenant, namespace, name)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *iamServer) CreateRole(g *gin.Context) {
	request := &rbac.Role{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	res, err := i.RoleService.Create(request)
	if err != nil {
		i.RecordEvent(common.ROLE, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.ROLE, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *iamServer) UpdateRole(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	tenant := g.DefaultQuery("tenant", "")
	paths := api.SplitPath(g.Query("path"))

	request := &rbac.Role{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := i.RoleService.Update(tenant, namespace, name, request, paths...)
	if err != nil {
		i.RecordEvent(common.ROLE, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.ROLE, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *iamServer) DeleteRole(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	res, err := i.RoleService.Delete(tenant, namespace, name)
	if err != nil {
		i.RecordEvent(common.ROLE, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.ROLE, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
