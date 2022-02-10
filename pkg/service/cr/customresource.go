package cr

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/cr"
	"github.com/ddx2x/oilmont/pkg/service"
)

type CustomResourceService struct {
	service.IService
}

func NewCustomResourceService(i service.IService) *CustomResourceService {
	return &CustomResourceService{i}
}

func (ss *CustomResourceService) List(name, workspace string) (*cr.CustomResourceList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	data := make([]cr.CustomResource, 0)
	err := ss.IService.ListToObject(common.DefaultDatabase, common.CUSTOMRESOURCE, filter, &data, true)
	if err != nil {
		return nil, err
	}

	customResourceList := &cr.CustomResourceList{Items: data}
	customResourceList.GenerateListVersion()

	return customResourceList, nil
}

func (ss *CustomResourceService) GetByName(workspace, name string) (*cr.CustomResource, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	customResource := &cr.CustomResource{}
	err := ss.IService.GetByFilter(common.DefaultDatabase, common.CUSTOMRESOURCE, customResource, filter, true)
	if err != nil {
		return nil, err
	}
	return customResource, nil
}

func (ss *CustomResourceService) Create(reqCustomResource *cr.CustomResource) (core.IObject, error) {
	if _, err := ss.GetByName(reqCustomResource.GetWorkspace(), reqCustomResource.GetName()); err != datasource.NotFound {
		return nil, fmt.Errorf("customResource exists")
	}

	if reqCustomResource.Name == "" {
		return nil, fmt.Errorf("data invalid name not define")
	}

	reqCustomResource.Kind = cr.CustomResourceKind

	reqCustomResource.GenerateVersion()

	_, err := ss.IService.Create(common.DefaultDatabase, common.CUSTOMRESOURCE, reqCustomResource)
	if err != nil {
		return nil, err
	}
	return reqCustomResource, nil
}

func (ss *CustomResourceService) Update(name string, reqCustomResource *cr.CustomResource) (core.IObject, bool, error) {
	customResource, err := ss.GetByName(reqCustomResource.GetWorkspace(), name)
	if err != nil {
		return nil, false, err
	}

	customResource.Spec = reqCustomResource.Spec
	customResource.Spec = cr.CustomResourceSpec{}
	_, _, err = ss.IService.Apply(common.DefaultDatabase, common.CUSTOMRESOURCE, customResource.Name, customResource, true)
	if err != nil {
		return nil, false, err
	}

	customResource.Spec = reqCustomResource.Spec
	_, update, err := ss.IService.Apply(common.DefaultDatabase, common.CUSTOMRESOURCE, customResource.Name, customResource, true)
	if err != nil {
		return nil, false, err
	}

	return customResource, update, nil
}

func (ss *CustomResourceService) Delete(workspace, name string) (core.IObject, error) {
	object, err := ss.GetByName(workspace, name)
	if err != nil {
		return nil, err
	}

	if err := ss.DeleteObject(common.DefaultDatabase, common.CUSTOMRESOURCE, name, object, true); err != nil {
		return nil, err
	}
	// and remove CUSTOMRESOURCE table
	if err = ss.RemoveTable(common.CustomDatabase, name); err != nil {
		return nil, err
	}

	return object, nil
}
