package service

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"time"
)

type IService interface {
	Create(db, resource string, object core.IObject) (core.IObject, error)
	DeleteObject(db, resource, name string, object core.IObject, purge bool) error
	Delete(db, resource, name, workspace string) error
	Apply(db, resource, name string, object core.IObject, forceApply bool, paths ...string) (core.IObject, bool, error)
	List(db, resource, labels string, filterDelete bool) ([]interface{}, error)
	Get(db, resource, name string, result interface{}, filterDelete bool) error

	GetByMetadataUUID(db, resource, uuid string, result interface{}, filterDelete bool) error
	GetByFilter(db, resource string, result interface{}, filter map[string]interface{}, filterDelete bool) error
	DeleteByUUID(db, resource, uuid string) error

	Watch(db, resource string, resourceVersion string, watch datasource.WatchInterface, filters ...datasource.Filter)
	WatchEvent(ctx context.Context, db, resource string, resourceVersion string, filters ...datasource.Filter) (<-chan core.Event, error)

	ListToObject(db, resource string, filter map[string]interface{}, result interface{}, filterDelete bool) error
	ListByFilter(db, resource string, filter map[string]interface{}, filterDelete bool) ([]interface{}, error)

	GetSelf() IService
	BatchLoad(db, resource string, objects []core.IObject, forceApply bool) error
	RemoveTable(db, table string) error

	//Cache
	GetCache(k string) (interface{}, bool)
	SetCache(k string, x interface{}, d time.Duration)

	//event
	Add(*event.CloudEvent)
}

type BaseService struct {
	datasource.IStorage
	datasource.ICache
}

func (bs *BaseService) Add(event *event.CloudEvent) {
	bs.Create(common.DefaultDatabase, common.CLOUDEVENT, event)
}

func (bs *BaseService) DeleteObject(db, resource, name string, object core.IObject, purge bool) error {
	if err := bs.DeleteByIObject(db, resource, object); err != nil {
		return err
	}
	return nil
}

func (bs *BaseService) BatchLoad(db, resource string, objects []core.IObject, forceApply bool) error {
	return bs.Bulk(db, resource, objects)
}

func (bs *BaseService) RemoveTable(db, table string) error {
	return bs.IStorage.RemoveTable(db, table)
}

func (bs *BaseService) GetSelf() IService { return bs }

func NewBaseService(s datasource.IStorage, c datasource.ICache) IService {
	return &BaseService{s, c}
}
