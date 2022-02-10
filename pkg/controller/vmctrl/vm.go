package vmctrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ controller.Handler = &VMCtrl{}

var (
	podGvr                    = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	dataVolumeGvr             = schema.GroupVersionResource{Group: "cdi.kubevirt.io", Version: "v1beta1", Resource: "datavolume"}
	virtualMachineGvr         = schema.GroupVersionResource{Group: "github.com/ddx2x", Version: "v1", Resource: "virtualmachines"}
	virtualMachineInstanceGvr = schema.GroupVersionResource{Group: "kubevirt.io", Version: "v1", Resource: "virtualmachineinstances"}
)

type VMCtrl struct {
	stage datasource.IStorage
	cs    *clients.Clients
	flog  log.Logger
}

func (V *VMCtrl) Set(cs *clients.Clients, stage datasource.IStorage) {
	V.cs, V.stage = cs, stage
}

func NewVMCtrl(ctx context.Context) controller.Handler {
	flog := log.GetLogger(ctx).WithField("controller", "vmctrl")
	return &VMCtrl{flog: flog}
}
