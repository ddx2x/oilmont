package zonectrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/core"
)

func (V ZoneCtrl) NorthOnAdd(obj core.IObject) {
	panic("implement me")
}

func (V ZoneCtrl) NorthOnUpdate(obj core.IObject) {
	panic("implement me")
}

func (V ZoneCtrl) NorthOnDelete(obj core.IObject) {
	panic("implement me")
}

func (V ZoneCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return nil, nil
}
