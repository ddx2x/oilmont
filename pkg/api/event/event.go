package event

import (
	"fmt"
	"time"

	"github.com/ddx2x/oilmont/pkg/service"
	"github.com/ddx2x/oilmont/pkg/service/event"

	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/cache"
	"github.com/ddx2x/oilmont/pkg/micro/webservice"
)

type eventServer struct {
	api.IAPIServer
	webServer webservice.Server

	cloudEvent *event.CloudEventService
}

func (e *eventServer) Run() error {
	return e.webServer.Run()
}

func NewEventServer(serviceName string, storage datasource.IStorage) (*eventServer, error) {
	c := cache.NewCache(15*time.Minute, 20*time.Minute)
	baseService := service.NewBaseService(storage, c)
	server := api.NewBaseAPIServer(baseService)

	eventObj := &eventServer{
		IAPIServer: server,
		cloudEvent: event.NewCloudEvent(server.IService),
	}

	webServer, err := webservice.NewWEBServer(serviceName, "", eventObj.Server())
	if err != nil {
		return nil, err
	}
	eventObj.webServer = webServer

	group := eventObj.Server().Group(fmt.Sprintf("/%s", serviceName))

	// instance
	{
		api.GenerateURIV2(group, "event.ddx2x.nip", "v1", "cloudevent", true,
			eventObj.ListCloudEvent,
			eventObj.GetCloudEvent,
			nil,
			nil,
			nil,
		)
	}

	return eventObj, nil
}
