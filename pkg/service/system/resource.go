package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/resource/system"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/service"
)

type ResourceService struct {
	service.IService
}

func NewResource(i service.IService) *ResourceService {
	return &ResourceService{i}
}

func (p *ResourceService) List(name string) (core.IObjectList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}

	data := make([]system.Resource, 0)
	err := p.IService.ListToObject(common.DefaultDatabase, common.RESOURCE, filter, &data, true)
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

func (p *ResourceService) GetByName(name string) (*system.Resource, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}

	ResourceObj := &system.Resource{}
	if err := p.IService.GetByFilter(common.DefaultDatabase, common.RESOURCE, ResourceObj, filter, true); err != nil {
		return nil, err
	}

	return ResourceObj, nil
}

func (p *ResourceService) Create(reqResource *system.Resource) (core.IObject, error) {
	if reqResource.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	if _, err := p.GetByName(reqResource.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("resource exists")
	}

	if reqResource.Spec.ApiVersion == "" || reqResource.Spec.Group == "" {
		return nil, fmt.Errorf("resource params is empty")
	}

	reqResource.Kind = system.ResourceKind
	reqResource.GenerateVersion()

	_, err := p.IService.Create(common.DefaultDatabase, common.RESOURCE, reqResource)
	if err != nil {
		return nil, err
	}
	return reqResource, nil
}

func (p *ResourceService) Update(name string, reqResource *system.Resource, paths ...string) (core.IObject, bool, error) {
	return p.IService.Apply(common.DefaultDatabase, common.RESOURCE, name, reqResource, false, paths...)
}

func (p *ResourceService) Delete(name string) (core.IObject, error) {
	object, err := p.GetByName(name)
	if err != nil {
		return nil, err
	}

	err = p.DeleteObject(common.DefaultDatabase, common.RESOURCE, name, object, true)
	return object, err
}
