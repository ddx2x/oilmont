package instancetypectrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ controller.Handler = &InstanceTypeCtrl{}

var (
	instanceTypeGvr = schema.GroupVersionResource{Group: "github.com/ddx2x", Version: "v1", Resource: "instancetypes"}
)

type InstanceTypeCtrl struct {
	flog  log.Logger
	stage datasource.IStorage
	cs    *clients.Clients
}

func (V *InstanceTypeCtrl) Set(cs *clients.Clients, stage datasource.IStorage) {
	V.cs, V.stage = cs, stage
}

func NewInstanceTypeCtrl(ctx context.Context) controller.Handler {
	flog := log.GetLogger(ctx).WithField("controller", "instanceTypeCtrl")
	return &InstanceTypeCtrl{flog: flog}
}
