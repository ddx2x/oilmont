package main

import (
	"context"
	"os"

	"github.com/ddx2x/oilmont/pkg/controller/iamctrl"
	"github.com/ddx2x/oilmont/pkg/datasource/mongo"
	"github.com/ddx2x/oilmont/pkg/log"
	logruslogger "github.com/ddx2x/oilmont/pkg/log/logrus"
	"github.com/ddx2x/oilmont/pkg/thirdparty/signals"
	"github.com/sirupsen/logrus"
)

var uri string
var DefaultStorageUrl = "mongodb://127.0.0.1:27017/admin"

func main() {
	stopCh := signals.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stopCh
		cancel()
	}()

	log.L = logruslogger.FromLogrus(logrus.NewEntry(logrus.StandardLogger()))
	log.G(ctx).Info("start iamctrl controller")

	uri = os.Getenv("STORAGE_URI")
	if uri == "" {
		uri = DefaultStorageUrl
	}

	stage, err, errC := mongo.NewMongo(ctx, uri)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := iamctrl.NewRBACController(stage).Run(); err != nil {
			errC <- err
		}
	}()

	panic(<-errC)

}
