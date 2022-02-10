package regionctrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/core"
)

func (V RegionCtrl) NorthOnAdd(obj core.IObject) {
	panic("implement me")
}

func (V RegionCtrl) NorthOnUpdate(obj core.IObject) {
	panic("implement me")
}

func (V RegionCtrl) NorthOnDelete(obj core.IObject) {
	panic("implement me")
}

func (V RegionCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return nil, nil
}
