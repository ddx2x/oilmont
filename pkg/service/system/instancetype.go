package system

import (
	"github.com/ddx2x/oilmont/pkg/resource/system"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/service"
)

type InstanceTypeService struct {
	service.IService
}

func NewInstanceType(i service.IService) *InstanceTypeService {
	return &InstanceTypeService{i}
}

func (is *InstanceTypeService) List(name, namespace, provider, region, zone string) (*system.InstanceTypeList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if provider != "" {
		// instanceType namespace field is provider
		filter[common.FilterNamespace] = provider
	}
	if region != "" {
		filter["spec.region"] = region
	}
	if zone != "" {
		filter["spec.zone"] = zone
	}

	data := make([]system.InstanceType, 0)
	err := is.IService.ListToObject(common.DefaultDatabase, common.INSTANCETYPE, filter, &data, true)
	if err != nil {
		return nil, err
	}

	instanceTypeList := &system.InstanceTypeList{Items: data}
	instanceTypeList.GenerateListVersion()

	return instanceTypeList, nil
}

func (is *InstanceTypeService) GetByName(namespace, name string) (*system.InstanceType, error) {
	filter := map[string]interface{}{
		common.FilterName:      name,
		common.FilterNamespace: namespace,
	}

	instanceType := &system.InstanceType{}
	err := is.IService.GetByFilter(common.DefaultDatabase, common.INSTANCETYPE, instanceType, filter, true)
	if err != nil {
		return nil, err
	}
	return instanceType, nil
}
