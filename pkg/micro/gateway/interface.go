package gateway

import (
	"github.com/micro/micro/v2/cmd"
	"github.com/micro/micro/v2/plugin"
	"github.com/ddx2x/oilmont/pkg/micro"
	"net/http"
)

type InterceptType uint8

const (
	SelfHandle InterceptType = iota
	Redirect
	Interrupt
	NotAuthorized
	Next
)

type Intercept func(w http.ResponseWriter, r *http.Request) InterceptType

type Server interface {
	micro.IMicroServer
}

func NewGatewayServer(handler http.Handler, intercepts ...Intercept) (Server, error) {
	handlerWrappers := []plugin.Handler{
		ServerIntercept(handler, intercepts...),
	}
	if err := plugin.Register(plugin.NewPlugin(plugin.WithHandler(handlerWrappers...))); err != nil {
		return nil, err
	}
	return &gatewayServer{}, nil
}

var _ Server = &gatewayServer{}

type gatewayServer struct{}

func (g gatewayServer) UUID() string {
	return ""
}

func (g gatewayServer) Run() error {
	cmd.Init()
	return nil
}
