package imagectrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/core"
)

func (V ImageCtrl) NorthOnAdd(obj core.IObject) {
	panic("implement me")
}

func (V ImageCtrl) NorthOnUpdate(obj core.IObject) {
	panic("implement me")
}

func (V ImageCtrl) NorthOnDelete(obj core.IObject) {
	panic("implement me")
}

func (V ImageCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return nil, nil
}
