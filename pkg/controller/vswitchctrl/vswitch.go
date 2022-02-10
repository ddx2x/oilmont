package vswitchctrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	vSwitchGvr = schema.GroupVersionResource{Group: "github.com/ddx2x", Version: "v1", Resource: "vswitches"}
)

var _ controller.Handler = &VSwitchCtrl{}

type VSwitchCtrl struct {
	stage datasource.IStorage
	cs    *clients.Clients
	flog  log.Logger
}

func (V *VSwitchCtrl) Set(cs *clients.Clients, stage datasource.IStorage) {
	V.cs, V.stage = cs, stage
}

func NewVSwitchCtrl(ctx context.Context) controller.Handler {
	flog := log.GetLogger(ctx).WithField("controller", "vSwitchCtrl")
	return &VSwitchCtrl{flog: flog}
}
