package instancetypectrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/core"
)

func (V InstanceTypeCtrl) NorthOnAdd(obj core.IObject) {
	panic("implement me")
}

func (V InstanceTypeCtrl) NorthOnUpdate(obj core.IObject) {
	panic("implement me")
}

func (V InstanceTypeCtrl) NorthOnDelete(obj core.IObject) {
	panic("implement me")
}

func (V InstanceTypeCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return nil, nil
}
