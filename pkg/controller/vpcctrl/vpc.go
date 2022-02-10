package vpcctrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	virtualPrivateCloudGvr = schema.GroupVersionResource{Group: "github.com/ddx2x", Version: "v1", Resource: "vpcs"}
)

var _ controller.Handler = &VPCCtrl{}

type VPCCtrl struct {
	stage datasource.IStorage
	cs    *clients.Clients
	flog  log.Logger
}

func (V *VPCCtrl) Set(cs *clients.Clients, stage datasource.IStorage) {
	V.cs, V.stage = cs, stage
}

func NewVPCtrl(ctx context.Context) controller.Handler {
	flog := log.GetLogger(ctx).WithField("controller", "vpcCtrl")
	return &VPCCtrl{flog: flog}
}
