package cr

import (
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/cr"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"net/http"
)

func (c *CustomResourceServer) ListCustomResource(g *gin.Context) {
	namespace := g.Param("namespace")
	results, err := c.customResource.List("", namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (c *CustomResourceServer) GetCustomResource(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")

	results, err := c.customResource.List(name, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (c *CustomResourceServer) CreateCustomResource(g *gin.Context) {
	request := &cr.CustomResource{}

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := c.customResource.Create(request)
	if err != nil {
		c.RecordEvent(common.CUSTOMRESOURCE, core.ADDED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	c.RecordEvent(common.CUSTOMRESOURCE, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (c *CustomResourceServer) UpdateCustomResource(g *gin.Context) {
	name := g.Param("name")
	request := &cr.CustomResource{}

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := c.customResource.Update(name, request)
	if err != nil {
		c.RecordEvent(common.CUSTOMRESOURCE, core.MODIFIED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	c.RecordEvent(common.CUSTOMRESOURCE, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (c *CustomResourceServer) DeleteCustomResource(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := c.customResource.Delete(namespace, name)
	if err != nil {
		c.RecordEvent(common.CUSTOMRESOURCE, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	c.RecordEvent(common.CUSTOMRESOURCE, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
