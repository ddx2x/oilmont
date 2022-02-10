package zonectrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ controller.Handler = &ZoneCtrl{}

var (
	zoneGvr = schema.GroupVersionResource{Group: "github.com/ddx2x", Version: "v1", Resource: "zones"}
)

type ZoneCtrl struct {
	flog  log.Logger
	stage datasource.IStorage
	cs    *clients.Clients
}

func (V *ZoneCtrl) Set(cs *clients.Clients, stage datasource.IStorage) {
	V.cs, V.stage = cs, stage
}

func NewZoneCtrl(ctx context.Context) controller.Handler {
	flog := log.GetLogger(ctx).WithField("controller", "zonectrl")
	return &ZoneCtrl{flog: flog}
}
