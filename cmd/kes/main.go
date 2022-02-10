package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ddx2x/oilmont/pkg/api/kes"
	"github.com/ddx2x/oilmont/pkg/datasource/mongo"
	"github.com/ddx2x/oilmont/pkg/k8s"
	"github.com/ddx2x/oilmont/pkg/log"
	logruslogger "github.com/ddx2x/oilmont/pkg/log/logrus"
	"github.com/ddx2x/oilmont/pkg/thirdparty/signals"
	"github.com/sirupsen/logrus"
)

const (
	DefaultStorageUrl = "mongodb://127.0.0.1:27017/admin"
	serviceName       = "kes"
)

func main() {
	stopCh := signals.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stopCh
		cancel()
	}()

	log.L = logruslogger.FromLogrus(logrus.NewEntry(logrus.StandardLogger()))
	log.G(context.Background()).Info("start kes webserver")

	uri := os.Getenv("STORAGE_URI")
	if uri == "" {
		uri = DefaultStorageUrl
	}

	stage, err, errC := mongo.NewMongo(ctx, uri)
	if err != nil {
		panic(fmt.Sprintf("init mongodb database connect error %s", err))
	}

	// add gvr to local data storage
	k8s.SetStorage(stage)

	server, err := kes.NewKesServer(serviceName, stage)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := server.Run(); err != nil {
			errC <- err
		}
	}()

	if e := <-errC; e != nil {
		panic(e)
	}

}
