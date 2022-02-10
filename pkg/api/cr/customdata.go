package cr

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/cr"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"github.com/gin-gonic/gin"
)

func (c *CustomResourceServer) ListCustomData(g *gin.Context) {
	namespace := g.Param("namespace")
	resource := g.Param("resource")

	results, err := c.customData.List(resource, "", namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (c *CustomResourceServer) GetCustomData(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")
	resource := g.Param("resource")

	results, err := c.customData.List(resource, name, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

type uploadRequest struct {
	Data string `json:"data"`
}

func (u *uploadRequest) toCDS(kind string) ([]cr.CustomData, error) {
	rd := bufio.NewReader(strings.NewReader(u.Data))
	index := 0
	cds := make([]cr.CustomData, 0)
	var headerPosition []string

	addLine := func(line string, index int) error {
		content := make(map[string]interface{})
		spec := make(map[string]interface{})
		meta := core.Metadata{
			Kind: core.Kind(kind),
		}

		lineList := strings.Split(line, ",")
		if len(lineList) == 0 {
			return fmt.Errorf("handle line %d Invaild symbol", index+1)
		}
		if index == 0 {
			// header
			headerPosition = lineList
			for _, header := range lineList {
				content[header] = nil
			}
			return nil
		} else {
			//data
			for i := range lineList {
				value := lineList[i]
				switch headerPosition[i] {
				case "id":
				case "name":
					meta.Name = value
				case "workspace":
					meta.Workspace = value
				default:
					spec[headerPosition[i]] = value
				}
			}
		}
		cd := cr.CustomData{
			Metadata: meta,
			Spec:     spec,
		}

		cd.GenerateVersion()
		if cd.GetName() == "" || cd.GetWorkspace() == "" {
			//ignore invalid data
			return nil
		}
		cds = append(cds, cd)
		return nil
	}

	for {
		line, err := rd.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if err != nil {
			if err == io.EOF {
				if rErr := addLine(line, index); rErr != nil {
					return nil, rErr
				}
				break
			}
			return nil, err
		}
		if rErr := addLine(line, index); rErr != nil {
			return nil, rErr
		}
		index++
	}
	return cds, nil
}

func (c *CustomResourceServer) upload(g *gin.Context) {
	resource := g.Param("resource")
	uploadRequest := &uploadRequest{}
	if err := g.BindJSON(uploadRequest); err != nil {
		api.RequestParametersError(g, err)
		return
	}
	cds, err := uploadRequest.toCDS(resource)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	err = c.customData.Upload(resource, cds)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, "")
}

func (c *CustomResourceServer) CreateCustomData(g *gin.Context) {
	customDatabase := "custom"
	resource := g.Param("resource")
	request := &cr.CustomData{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	res, err := c.customData.Create(resource, request)
	if err != nil {
		c.RecordEvent(customDatabase, core.ADDED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	c.RecordEvent(customDatabase, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (c *CustomResourceServer) UpdateCustomData(g *gin.Context) {
	customDatabase := "custom"
	namespace := g.Param("namespace")
	name := g.Param("name")
	resource := g.Param("resource")

	request := &cr.CustomData{}

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, _, err := c.customData.Update(resource, namespace, name, request)
	if err != nil {
		c.RecordEvent(customDatabase, core.MODIFIED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	c.RecordEvent(customDatabase, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (c *CustomResourceServer) DeleteCustomData(g *gin.Context) {
	customDatabase := "custom"
	workspace := g.Param("namespace")
	name := g.Param("name")
	resource := g.Param("resource")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	res, err := c.customData.Delete(resource, workspace, name)
	if err != nil {
		c.RecordEvent(customDatabase, core.DELETED, reqUser, res, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	c.RecordEvent(customDatabase, core.DELETED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}
