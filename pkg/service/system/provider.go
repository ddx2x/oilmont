package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
	"reflect"
)

type ProviderService struct {
	service.IService
}

func NewProviderService(i service.IService) *ProviderService {
	return &ProviderService{i}
}

func (rs *ProviderService) List(name string) (core.IObjectList, error) {
	filter := make(map[string]interface{})
	if name != "" {
		filter[common.FilterName] = name
	}
	data := make([]system.Provider, 0)
	err := rs.IService.ListToObject(common.DefaultDatabase, common.PROVIDER, filter, &data, true)
	if err != nil {
		return nil, err
	}

	items := make([]core.IObject, 0)
	for _, item := range data {
		itemP := item
		items = append(items, core.ToItems(&itemP)...)
	}
	return core.NewIObjectList(items), nil
}

func (rs *ProviderService) GetByName(name string) (*system.Provider, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}
	object := &system.Provider{}
	err := rs.IService.GetByFilter(common.DefaultDatabase, common.PROVIDER, object, filter, true)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (rs *ProviderService) Create(object *system.Provider) (core.IObject, error) {
	if _, err := rs.GetByName(object.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("object exists")
	}

	if object.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	object.Kind = system.PorviderKind
	object.GenerateVersion()

	_, err := rs.IService.Create(common.DefaultDatabase, common.PROVIDER, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (rs *ProviderService) Update(name string, object *system.Provider) (core.IObject, bool, error) {
	getObject, err := rs.GetByName(name)
	if err != nil {
		return nil, false, err
	}

	if !reflect.DeepEqual(getObject.Spec, object.Spec) {
		getObject.Spec = object.Spec
	}

	_, update, err := rs.IService.Apply(common.DefaultDatabase, common.PROVIDER, getObject.Name, getObject, false)
	if err != nil {
		return nil, false, err
	}

	return getObject, update, nil
}

func (rs *ProviderService) Delete(name string) (core.IObject, error) {
	object, err := rs.GetByName(name)
	if err != nil {
		return nil, err
	}
	object.Delete()
	_, _, err = rs.Apply(common.DefaultDatabase, common.PROVIDER, object.Name, object, true)
	return object, err
}
