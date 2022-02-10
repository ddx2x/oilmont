package iam

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
)

func (i *iamServer) ListUser(g *gin.Context) {
	namespace := g.Param("namespace")
	tenantName := g.Query("tenant_name")
	// if user is admin use tenantName query
	if tenantName == "" {
		tenantName = g.GetHeader(common.HttpRequestUserHeaderTENANT)
	}
	results, err := i.UserService.List(tenantName, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, results)
}

func (i *iamServer) GetUser(g *gin.Context) {
	name := g.Param("name")
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)
	results, err := i.UserService.List(tenant, name)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *iamServer) UpdateUser(g *gin.Context) {
	tenant := g.DefaultQuery("tenant", "")

	request := &iam.User{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := i.UserService.Update(tenant, request)
	if err != nil {
		i.RecordEvent(common.USER, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.USER, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
