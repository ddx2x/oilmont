package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/cache"
	"github.com/ddx2x/oilmont/pkg/k8s"
	"github.com/ddx2x/oilmont/pkg/micro/gateway"
	third "github.com/ddx2x/oilmont/pkg/utils/thirdlogin"
	"github.com/ddx2x/oilmont/pkg/utils/thirdlogin/feishu"
	"github.com/ddx2x/oilmont/pkg/utils/uri"
	"github.com/gin-gonic/gin"
)

const (
	LoginURL       = "/user-login"
	FeiShuLoginURL = "/feishu-user-login"
	WatchURL       = "/watch"
	SHELL          = "/kes/shell/pod"
)

type Gateway struct {
	api.IAPIServer
	parser uri.Parser
	stage  datasource.IStorage
	perm   *permission
	third  third.IThirdPartLogin
	cache  datasource.ICache

	mc *k8s.MultiCluster
}

func NewGateway(stage datasource.IStorage) (*Gateway, error) {
	gw := &Gateway{
		IAPIServer: api.NewBaseAPIServer(nil),
		parser:     uri.NewURIParser(),
		stage:      stage,
		perm:       newPermission(stage),
		third:      feishu.NewFeiShu(),
		cache:      cache.NewCache(60*time.Minute, 70*time.Minute),

		mc: k8s.NewMultiCluster(stage),
	}

	server := gw.Server()
	server.Use(CORSMiddleware()) // 开发环境需要重转发sse/websocket 需要开启允许跨域。

	server.POST(LoginURL, gw.Login)
	server.POST(FeiShuLoginURL, gw.Login)
	server.GET("/watch", gw.watch)

	if err := gw.mc.AsyncRun(context.Background()); err != nil {
		return nil, err
	}

	return gw, nil
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
func (gw *Gateway) PermissionIntercept() gateway.Intercept {
	return gw.perm.Intercept()
}
