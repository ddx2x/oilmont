package networkinterfacectrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	NetworkInterfaceGvr = schema.GroupVersionResource{Group: "github.com/ddx2x", Version: "v1", Resource: "networkinterfaces"}
)

var _ controller.Handler = &NetworkInterfaceCtrl{}

type NetworkInterfaceCtrl struct {
	stage datasource.IStorage
	cs    *clients.Clients
	flog  log.Logger
}

func (V *NetworkInterfaceCtrl) Set(cs *clients.Clients, stage datasource.IStorage) {
	V.cs, V.stage = cs, stage
}

func NewNetworkInterfaceCtrl(ctx context.Context) controller.Handler {
	flog := log.GetLogger(ctx).WithField("controller", "networkInterfaceCtrl")
	return &NetworkInterfaceCtrl{flog: flog}
}
