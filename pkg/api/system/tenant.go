package system

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"net/http"
)

func (i *systemServer) ListTenant(g *gin.Context) {
	results, err := i.tenant.List("")
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateTenant(g *gin.Context) {
	request := &system.ReqTenant{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.tenant.Create(request)
	if err != nil {
		i.RecordEvent(common.TENANT, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.TENANT, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateTenant(g *gin.Context) {
	name := g.Param("name")
	path := g.Query("path")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	request := &system.Tenant{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	if path == "" {
		i.RecordEvent(common.TENANT, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, fmt.Errorf("not define path update"))
		return
	}

	res, _, err := i.tenant.Update(name, request, path)
	if err != nil {
		i.RecordEvent(common.TENANT, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.TENANT, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) DeleteTenant(g *gin.Context) {
	name := g.Param("name")
	res, err := i.tenant.Delete(name)
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	if err != nil {
		i.RecordEvent(common.TENANT, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.TENANT, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
