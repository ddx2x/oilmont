package cr

import (
	"context"
	"fmt"
	"time"

	"github.com/ddx2x/oilmont/pkg/service"

	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/cache"
	"github.com/ddx2x/oilmont/pkg/micro/webservice"
	customresource "github.com/ddx2x/oilmont/pkg/resource/cr"
	"github.com/ddx2x/oilmont/pkg/service/cr"
)

type CustomResourceServer struct {
	api.IAPIServer
	storage        datasource.IStorage
	webServer      webservice.Server
	customResource *cr.CustomResourceService
	customData     *cr.CustomDataService
}

func (c *CustomResourceServer) Run() error {
	go c.watchCustomResource()
	return c.webServer.Run()
}

func NewCustomResourceServer(serviceName string, storage datasource.IStorage) (*CustomResourceServer, error) {
	c := cache.NewCache(15*time.Minute, 20*time.Minute)
	baseService := service.NewBaseService(storage, c)
	server := api.NewBaseAPIServer(baseService)

	crs := &CustomResourceServer{
		IAPIServer: server,
		storage:    storage,

		customResource: cr.NewCustomResourceService(server.IService),
		customData:     cr.NewCustomData(server.IService),
	}

	webServer, err := webservice.NewWEBServer(serviceName, "", crs.Server())
	if err != nil {
		return nil, err
	}
	crs.webServer = webServer

	group := crs.Server().Group(fmt.Sprintf("/%s", serviceName))

	// customResource
	{
		api.GenerateURIV2(
			group, "cr.ddx2x.nip", "v1", "customresource", false,
			crs.ListCustomResource,
			crs.ListCustomResource,
			crs.CreateCustomResource,
			crs.UpdateCustomResource,
			crs.DeleteCustomResource,
		)
	}

	// CustomData
	{
		api.GenerateURIV2(
			group, "cr.ddx2x.nip", "v1", ":resource", true,
			crs.ListCustomData,
			crs.ListCustomData,
			crs.CreateCustomData,
			crs.UpdateCustomData,
			crs.DeleteCustomData,
		)

		group.POST("/apis/cr.ddx2x.nip/v1/:resource/op/upload", crs.upload)
	}

	return crs, nil
}

func (c *CustomResourceServer) watchCustomResource() {
	customDatabase := "custom"
	customResourceEvent, err := c.customResource.WatchEvent(context.Background(), common.DefaultDatabase, common.CUSTOMRESOURCE, "0")
	if err != nil {
		panic(err)
	}
	for {
		select {
		case event, ok := <-customResourceEvent:
			if !ok {
				return
			}
			crName := event.Object.GetName()
			switch event.Type {
			case core.ADDED:
				datasource.RegistryCoder(crName, &customresource.CustomData{})
				if err := common.InsertDynCR(c.storage, crName, customDatabase); err != nil {
					panic(err)
				}
			case core.DELETED:
				datasource.UNRegistryCoder(crName)
			}
		}
	}
}
