package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ddx2x/oilmont/pkg/k8s"
	"github.com/ddx2x/oilmont/pkg/thirdparty/signals"

	apiGateway "github.com/ddx2x/oilmont/pkg/api/gateway"
	"github.com/ddx2x/oilmont/pkg/datasource/mongo"
	"github.com/ddx2x/oilmont/pkg/log"
	logruslogger "github.com/ddx2x/oilmont/pkg/log/logrus"
	"github.com/ddx2x/oilmont/pkg/micro/gateway"
	"github.com/sirupsen/logrus"
)

var DefaultStorageUrl = "mongodb://127.0.0.1:27017/admin"
var uri string

func main() {
	stopCh := signals.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stopCh
		cancel()
	}()

	std := logrus.StandardLogger()
	std.SetLevel(logrus.DebugLevel)
	log.L = logruslogger.FromLogrus(logrus.NewEntry(std))
	log.G(ctx).Info("start gateway")

	uri = os.Getenv("STORAGE_URI")
	if uri == "" {
		uri = DefaultStorageUrl
	}

	stage, err, errC := mongo.NewMongo(ctx, uri)
	if err != nil {
		panic(fmt.Sprintf("init mongodb database connect error %s", err))
	}

	// add gvr to local data storage
	k8s.SetStorage(stage)

	gw, err := apiGateway.NewGateway(stage)
	if err != nil {
		panic(err)
	}

	server, err := gateway.NewGatewayServer(
		// 权限校验，开发过程不要打开 中间件执行具有先后顺序
		gw.Handler(), // gateway self handler
		apiGateway.LoginHandle,
		apiGateway.JWTToken,
		//apiGateway.DefaultRedirect,
		// 权限校验，开发过程不要打开
		gw.PermissionIntercept(),
	)

	if err != nil {
		panic(err)
	}

	go func() {
		if err = server.Run(); err != nil {
			errC <- err
		}
	}()

	for err := range errC {
		panic(err)
	}
}
