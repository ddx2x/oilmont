package k8s

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"sort"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

type GVR = schema.GroupVersionResource

var rr = &resourceRegistry{data: make(map[string]GVR)}
var ShardingResourceRegistry ResourceRegistry = rr

func SetStorage(stage datasource.IStorage) { rr.stage = stage }

type resourceRegistry struct {
	data  map[string]GVR
	stage datasource.IStorage
}

func (m *resourceRegistry) Register(res string, gvr GVR) { m.data[res] = gvr }

func (m *resourceRegistry) Subscript(d dynamicinformer.DynamicSharedInformerFactory, stop <-chan struct{}) {
	for res, gvr := range m.data {
		go d.ForResource(gvr).Informer().Run(stop)
		if err := m.stage.InsertUnique(common.DefaultDatabase, common.GVRRESOURCE, res, gvr); err != nil {
			panic(err)
		}
	}
	d.Start(stop)
}

type gvrTemp struct {
	Id   string `json:"_id" bson:"_id"`
	Data GVR    `json:"data" bson:"data"`
}

func (m *resourceRegistry) GetGVR(s string) (GVR, error) {
	item, exist := m.data[s]
	if exist && (item.Resource == s) {
		return item, nil
	}
	gvrTemp := &gvrTemp{}
	if err := m.stage.GetById(common.DefaultDatabase, common.GVRRESOURCE, s, gvrTemp); err != nil {
		return GVR{}, err
	}
	if gvrTemp.Data.Resource != s {
		return item, fmt.Errorf("resource not found %s", s)
	}
	gvr := GVR{Group: gvrTemp.Data.Group, Resource: gvrTemp.Data.Resource, Version: gvrTemp.Data.Version}
	m.data[s] = gvr

	return gvr, nil
}

func (m *resourceRegistry) Length() int {
	return len(m.data)
}

var _ sort.Interface = Resources{}

type Resources []Resource

func (r Resources) Len() int {
	return len(r)
}

func (r Resources) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

func (r Resources) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Resources) Strings() []string {
	result := make([]string, 0)
	for _, item := range r {
		result = append(result, item.Name)
	}
	return result
}

func (r Resources) In(x string) bool {
	if sort.SearchStrings(r.Strings(), x) > len(r.Strings())-1 {
		return false
	}
	return true
}
