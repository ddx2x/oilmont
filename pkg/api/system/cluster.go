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

func (i *systemServer) ListCluster(g *gin.Context) {
	results, err := i.cluster.List("")
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateCluster(g *gin.Context) {
	request := &system.Cluster{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.cluster.Create(common.DefaultDatabase, common.CLUSTER, request)
	if err != nil {
		i.RecordEvent(common.CLUSTER, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.CLUSTER, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateCluster(g *gin.Context) {
	name := g.Param("name")
	request := &system.Cluster{}

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := i.cluster.Update(name, request)
	if err != nil {
		i.RecordEvent(common.CLUSTER, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.CLUSTER, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) DeleteCluster(g *gin.Context) {
	name := g.Param("name")
	res, err := i.cluster.Delete(name)
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	if err != nil {
		i.RecordEvent(common.CLUSTER, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.CLUSTER, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
