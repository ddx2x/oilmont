package main

import (
	"context"
	"os"

	"github.com/ddx2x/oilmont/pkg/api/iam"
	"github.com/ddx2x/oilmont/pkg/datasource/mongo"
	"github.com/ddx2x/oilmont/pkg/log"
	logruslogger "github.com/ddx2x/oilmont/pkg/log/logrus"
	"github.com/ddx2x/oilmont/pkg/thirdparty/signals"
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

	log.L = logruslogger.FromLogrus(logrus.NewEntry(logrus.StandardLogger()))
	log.G(context.Background()).Info("start iam webserver")

	uri = os.Getenv("STORAGE_URI")
	if uri == "" {
		uri = DefaultStorageUrl
	}

	store, err, errC := mongo.NewMongo(ctx, uri)
	if err != nil {
		panic(err)
	}

	server, err := iam.NewIAMServer("iam", store)
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
