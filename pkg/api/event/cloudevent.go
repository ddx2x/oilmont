package event

import (
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"net/http"
)

func (e *eventServer) ListCloudEvent(g *gin.Context) {
	namespace := g.Param("namespace")
	results, err := e.cloudEvent.List("", namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (e *eventServer) GetCloudEvent(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")

	results, err := e.cloudEvent.List(name, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}
