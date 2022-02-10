package webservice

import (
	"fmt"
	"github.com/micro/go-micro/v2/web"
	"github.com/ddx2x/oilmont/pkg/micro"
	"net/http"
	"time"
)

const webNormalName = "go.micro.api.%s"

type Server interface {
	micro.IMicroServer
}

func NewWEBServer(name, version string, handler http.Handler) (Server, error) {
	webService := web.NewService(
		web.Name(fmt.Sprintf(webNormalName, name)),
		web.Version(version),
		web.RegisterTTL(time.Second*15),
		web.RegisterInterval(time.Second*10),
	)
	if err := webService.Init(); err != nil {
		return nil, err
	}
	webService.Handle("/", handler)

	return &webServer{webService}, nil
}

var _ Server = &webServer{}

type webServer struct {
	web.Service
}

func (w webServer) UUID() string {
	return w.Service.Options().Id
}
