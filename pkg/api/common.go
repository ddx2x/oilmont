package api

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	message = "message"
	data    = "data"
	errors  = "errors"
)

func LoginError(g *gin.Context) {
	content := `username does not exist or password is wrong`
	g.JSON(http.StatusBadRequest,
		gin.H{
			data:    content,
			message: content,
			errors:  content,
		},
	)
	g.Abort()
}

func SplitPath(path string) []string {
	cnt := strings.Count(path, ",")
	if cnt == 0 {
		return []string{path}
	}
	return strings.Split(path, ",")
}

func RequestParametersError(g *gin.Context, err error) {
	g.JSON(http.StatusBadRequest,
		gin.H{data: err.Error(), message: err.Error(), errors: err.Error()},
	)
	g.Abort()
}

func QueryFilter(g *gin.Context) (tenant string, filter map[string]interface{}) {
	if namespace := g.Param("namespace"); namespace != "" {
		filter[common.FilterNamespace] = namespace
	}
	if name := g.Param("name"); name != "" {
		filter[common.FilterName] = name
	}
	tenant = g.Query("tenant")
	return
}

func ResponseSuccess(g *gin.Context, data interface{}, msg string) {
	g.JSON(http.StatusOK, gin.H{message: msg, "code": http.StatusOK, "data": data})
	g.Abort()
}

func InternalServerError(g *gin.Context, _data interface{}, err error) {
	g.JSON(http.StatusInternalServerError,
		gin.H{data: _data, message: err.Error(), errors: err.Error()},
	)
	g.Abort()
}

func Unauthorized(g *gin.Context, _data interface{}) {
	g.JSON(http.StatusUnauthorized,
		gin.H{data: _data, message: "user unauthorized"},
	)
	g.Abort()
}

func Forbidden(g *gin.Context, _data interface{}) {
	g.JSON(http.StatusForbidden,
		gin.H{data: _data, message: "not allow to access"},
	)
	g.Abort()
}

func GenerateURI(apiVersion, group, resource string, namespaces bool) string {
	if namespaces {
		return fmt.Sprintf("/apis/%s/%s/namespaces/:namespace/%s", apiVersion, group, resource)
	}
	return fmt.Sprintf("/apis/%s/%s/%s", apiVersion, group, resource)
}

func GenerateURIV2(ginGroup *gin.RouterGroup, apiVersion, group, resource string, namespaces bool, listHandle, getHandle, createHandle, updateHandle, deleteHandle gin.HandlerFunc) {
	singleResource := fmt.Sprintf("%s/:name", resource)
	if getHandle != nil {
		ginGroup.GET(GenerateURI(apiVersion, group, singleResource, namespaces), listHandle)
	}

	if listHandle != nil {
		ginGroup.GET(GenerateURI(apiVersion, group, resource, false), listHandle)
		if namespaces {
			ginGroup.GET(GenerateURI(apiVersion, group, resource, namespaces), listHandle)
		}
	}

	if createHandle != nil {
		if namespaces {
			ginGroup.POST(GenerateURI(apiVersion, group, resource, namespaces), createHandle)
		}
		ginGroup.POST(GenerateURI(apiVersion, group, resource, false), createHandle)
	}

	if updateHandle != nil {
		if namespaces {
			ginGroup.PUT(GenerateURI(apiVersion, group, singleResource, namespaces), updateHandle)
		}
		ginGroup.PUT(GenerateURI(apiVersion, group, singleResource, false), updateHandle)
	}

	if deleteHandle != nil {
		if namespaces {
			ginGroup.DELETE(GenerateURI(apiVersion, group, singleResource, namespaces), deleteHandle)
		}
		ginGroup.DELETE(GenerateURI(apiVersion, group, singleResource, false), deleteHandle)
	}
}
