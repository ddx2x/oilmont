package regionctrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ controller.Handler = &RegionCtrl{}

var (
	regionGvr = schema.GroupVersionResource{Group: "github.com/ddx2x", Version: "v1", Resource: "regions"}
)

type RegionCtrl struct {
	stage datasource.IStorage
	cs    *clients.Clients
	flog  log.Logger
}

func (V *RegionCtrl) Set(cs *clients.Clients, stage datasource.IStorage) {
	V.cs, V.stage = cs, stage
}

func NewRegionCtrl(ctx context.Context) controller.Handler {
	flog := log.GetLogger(ctx).WithField("controller", "regionctrl")
	return &RegionCtrl{flog: flog}
}
