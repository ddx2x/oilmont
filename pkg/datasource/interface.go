package datasource

import (
	"context"
	"fmt"
	"github.com/ddx2x/oilmont/pkg/core"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

type ErrorType error

var NotFound ErrorType = fmt.Errorf("notFound")

var coderList = make(map[string]Coder)

func RegistryCoder(res string, coder Coder) { coderList[res] = coder }
func UNRegistryCoder(res string)            { delete(coderList, res) }

func GetCoder(res string) Coder { return coderList[res] }

type Coder interface {
	Decode(map[string]interface{}) (core.IObject, error)
}

type WatchInterface interface {
	ResultChan() <-chan core.IObject
	Handle(map[string]interface{}) error
	ErrorStop() chan error
	CloseStop() chan struct{}
}

type Watch struct {
	r     chan core.IObject
	err   chan error
	c     chan struct{}
	coder Coder
}

func NewWatch(coder Coder) *Watch {
	return &Watch{
		r:     make(chan core.IObject, 1),
		err:   make(chan error),
		c:     make(chan struct{}),
		coder: coder,
	}
}

func (w *Watch) Handle(opData map[string]interface{}) error {
	obj, err := w.coder.Decode(opData)
	if err != nil {
		return err
	}
	w.r <- obj
	return nil
}

func (w *Watch) ResultChan() <-chan core.IObject { return w.r }

func (w *Watch) CloseStop() chan struct{} { return w.c }

func (w *Watch) ErrorStop() chan error { return w.err }

type Filter struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type IWatchEvent interface {
	WatchEvent(ctx context.Context, db, table string, resourceVersion string, filter ...Filter) (<-chan core.Event, error)
}

type IStorage interface {
	Create(db, table string, object core.IObject) (core.IObject, error)
	Delete(db, table, name, workspace string) error
	DeleteByIObject(db, table string, object core.IObject) error
	Apply(db, table, name string, object core.IObject, forceApply bool, paths ...string) (core.IObject, bool, error)
	List(db, table, labels string, filterDelete bool) ([]interface{}, error)
	Get(db, table, name string, result interface{}, filterDelete bool) error

	GetByMetadataUUID(db, table, uuid string, result interface{}, filterDelete bool) error
	GetByFilter(db, table string, result interface{}, filter map[string]interface{}, filterDelete bool) error
	DeleteByUUID(db, table, uuid string) error

	ListToObject(db, table string, filter map[string]interface{}, result interface{}, filterDelete bool) error
	ListByFilter(db, table string, filter map[string]interface{}, filterDelete bool) ([]interface{}, error)
	Watch(db, table string, resourceVersion string, watch WatchInterface, filters ...Filter)

	InsertUnique(db, table string, id interface{}, data interface{}) error
	GetById(db, table, id string, result interface{}) error
	Bulk(db, table string, objects []core.IObject) error
	RemoveTable(db, table string) error

	IWatchEvent
}

type ICache interface {
	GetCache(k string) (interface{}, bool)
	SetCache(k string, x interface{}, d time.Duration)
}

type IDataSource interface {
	List(db string, gvr schema.GroupVersionResource, flag string, pos, size int64, selector interface{}) (*unstructured.UnstructuredList, error)
	Get(db, name string, gvr schema.GroupVersionResource, subresources ...string) (*unstructured.Unstructured, error)
	Apply(db, name string, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, forceUpdate bool) (*unstructured.Unstructured, bool, error)
	Delete(db, name string, gvr schema.GroupVersionResource) error
	Watch(db string, table, resourceVersion string, timeoutSeconds int64, selector interface{}) (<-chan watch.Event, error)
}
