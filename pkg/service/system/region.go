package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type RegionService struct {
	service.IService
}

func NewRegionService(i service.IService) *RegionService {
	return &RegionService{i}
}

func (rs *RegionService) List(name, namespace string) (*system.RegionList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if namespace != "" {
		filter[common.FilterNamespace] = namespace
	}

	data := make([]system.Region, 0)
	err := rs.IService.ListToObject(common.DefaultDatabase, common.REGION, filter, &data, true)
	if err != nil {
		return nil, err
	}

	regionList := &system.RegionList{Items: data}
	regionList.GenerateListVersion()

	return regionList, nil
}

func (rs *RegionService) GetByName(namespace, name string) (*system.Region, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if namespace != "" {
		filter[common.FilterNamespace] = namespace
	}

	region := &system.Region{}
	err := rs.IService.GetByFilter(common.DefaultDatabase, common.REGION, region, filter, true)
	if err != nil {
		return nil, err
	}

	return region, nil
}

func (rs *RegionService) Create(reqRegion *system.Region) (core.IObject, error) {
	if _, err := rs.GetByName(reqRegion.Workspace, reqRegion.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("region exists")
	}

	if reqRegion.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	reqRegion.Kind = system.RegionKind
	reqRegion.GenerateVersion()

	_, err := rs.IService.Create(common.DefaultDatabase, common.REGION, reqRegion)
	if err != nil {
		return nil, err
	}
	return reqRegion, nil
}

func (rs *RegionService) Update(namespace, name string, reqRegion *system.Region) (core.IObject, bool, error) {
	region, err := rs.GetByName(namespace, name)
	if err != nil {
		return nil, false, err
	}

	_, update, err := rs.IService.Apply(common.DefaultDatabase, common.REGION, region.Name, region, false)
	if err != nil {
		return nil, false, err
	}

	return region, update, nil
}

func (rs *RegionService) Delete(name string) (core.IObject, error) {
	region, err := rs.GetByName("", name)
	if err != nil {
		return nil, err
	}
	err = rs.DeleteObject(common.DefaultDatabase, common.REGION, region.Name, region, true)

	return region, err
}
