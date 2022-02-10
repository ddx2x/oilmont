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

func (i *systemServer) ListTheme(g *gin.Context) {
	results, err := i.themeService.List("")
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, results)
}

func (i *systemServer) CreateTheme(g *gin.Context) {
	request := &system.Theme{}

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.themeService.Create(request)
	if err != nil {
		i.RecordEvent(common.THEME, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.THEME, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) UpdateTheme(g *gin.Context) {
	name := g.Param("name")

	request := &system.Theme{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := i.themeService.Update(name, request)
	if err != nil {
		i.RecordEvent(common.THEME, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.THEME, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *systemServer) DeleteTheme(g *gin.Context) {
	name := g.Param("name")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := i.themeService.Delete(name)
	if err != nil {
		i.RecordEvent(common.THEME, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.THEME, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
