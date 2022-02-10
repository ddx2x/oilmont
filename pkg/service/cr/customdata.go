package cr

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/cr"
	"github.com/ddx2x/oilmont/pkg/service"
	"strings"
)

type CustomDataService struct {
	service.IService
}

func NewCustomData(i service.IService) *CustomDataService {
	return &CustomDataService{i}
}

func (ss *CustomDataService) List(resource, name, workspace string) (*cr.CustomDataList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	data := make([]cr.CustomData, 0)
	err := ss.IService.ListToObject(common.CustomDatabase, resource, filter, &data, true)
	if err != nil {
		return nil, err
	}

	customResourceList := &cr.CustomDataList{Items: data}
	customResourceList.GenerateListVersion()

	return customResourceList, nil
}

func (ss *CustomDataService) GetByName(resource, workspace, name string) (*cr.CustomData, error) {
	filter := map[string]interface{}{
		common.FilterName:      name,
		common.FilterWorkspace: workspace,
	}

	customData := &cr.CustomData{}
	err := ss.IService.GetByFilter(common.CustomDatabase, resource, customData, filter, true)
	if err != nil {
		return nil, err
	}
	return customData, nil
}

func (ss *CustomDataService) Create(resource string, reqCustomData *cr.CustomData) (core.IObject, error) {
	if _, err := ss.GetByName(resource, reqCustomData.GetWorkspace(), reqCustomData.GetName()); err != datasource.NotFound {
		return nil, fmt.Errorf("customData exists")
	}

	if reqCustomData.Name == "" || reqCustomData.Workspace == "" {
		return nil, fmt.Errorf("data is empty")
	}
	reqCustomData.Kind = core.Kind(strings.ToLower(resource))

	reqCustomData.GenerateVersion()

	_, err := ss.IService.Create(common.CustomDatabase, resource, reqCustomData)
	if err != nil {
		return nil, err
	}
	return reqCustomData, nil
}

func (ss *CustomDataService) Upload(resource string, reqCustomData []cr.CustomData) error {
	objects := make([]core.IObject, len(reqCustomData))
	for index := range reqCustomData {
		objects[index] = &reqCustomData[index]
	}
	err := ss.IService.BatchLoad(common.CustomDatabase, resource, objects, false)
	if err != nil {
		return err
	}
	return nil
}

func (ss *CustomDataService) Update(resource, namespace, name string, reqCustomData *cr.CustomData) (core.IObject, bool, error) {
	customData, err := ss.GetByName(resource, namespace, name)
	if err != nil {
		return nil, false, err
	}

	customData.Spec = reqCustomData.Spec
	customData.Spec = nil
	_, _, err = ss.IService.Apply(common.CustomDatabase, resource, customData.Name, customData, true)
	if err != nil {
		return nil, false, err
	}

	customData.Spec = reqCustomData.Spec
	_, update, err := ss.IService.Apply(common.CustomDatabase, resource, customData.Name, customData, true)
	if err != nil {
		return nil, false, err
	}

	return customData, update, nil
}

func (ss *CustomDataService) Delete(resource, workspace, name string) (core.IObject, error) {
	object, err := ss.GetByName(resource, workspace, name)
	if err != nil {
		return nil, err
	}

	if err := ss.DeleteObject(common.CustomDatabase, resource, object.GetName(), object, true); err != nil {
		return nil, err
	}

	return object, nil
}
