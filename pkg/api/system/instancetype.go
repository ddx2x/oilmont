package system

import (
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"net/http"
)

func (i *systemServer) ListInstanceType(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")
	provider := g.Query("provider")
	region := g.Query("region")
	zone := g.Query("zone")

	results, err := i.instanceType.List(name, namespace, provider, region, zone)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}
