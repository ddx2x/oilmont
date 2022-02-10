package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type MenuService struct {
	service.IService
}

func NewMenuService(i service.IService) *MenuService {
	return &MenuService{i}
}

func (rs *MenuService) List() (core.IObjectList, error) {
	data := make([]system.Menu, 0)
	err := rs.IService.ListToObject(common.DefaultDatabase, common.Menu, nil, &data, true)
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

func (rs *MenuService) GetByName(name string) (*system.Menu, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}
	object := &system.Menu{}
	err := rs.IService.GetByFilter(common.DefaultDatabase, common.Menu, object, filter, true)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (rs *MenuService) Create(object *system.Menu) (core.IObject, error) {
	if _, err := rs.GetByName(object.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("object exists")
	}

	if object.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	object.Kind = system.MenuKind
	object.GenerateVersion()

	_, err := rs.IService.Create(common.DefaultDatabase, common.Menu, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (rs *MenuService) Update(name string, object *system.Menu, path string) (core.IObject, bool, error) {
	return rs.IService.Apply(common.DefaultDatabase, common.Menu, name, object, false, []string{path}...)
}

func (rs *MenuService) Delete(name string) (core.IObject, error) {
	object, err := rs.GetByName(name)
	if err != nil {
		return nil, err
	}
	err = rs.DeleteObject(common.DefaultDatabase, common.Menu, object.Name, object, true)
	return object, err
}
