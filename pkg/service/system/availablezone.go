package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type AvailableZoneService struct {
	service.IService
}

func NewAvailableZoneService(i service.IService) *AvailableZoneService {
	return &AvailableZoneService{i}
}

func (as *AvailableZoneService) List(name, namespace string) (*system.AvailableZoneList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if namespace != "" {
		filter[common.FilterNamespace] = namespace
	}

	data := make([]system.AvailableZone, 0)
	err := as.IService.ListToObject(common.DefaultDatabase, common.AVAILABLEZONE, filter, &data, true)
	if err != nil {
		return nil, err
	}

	availableZoneList := &system.AvailableZoneList{Items: data}
	availableZoneList.GenerateListVersion()

	return availableZoneList, nil
}

func (as *AvailableZoneService) GetByName(name string) (*system.AvailableZone, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}

	availableZone := &system.AvailableZone{}
	err := as.IService.GetByFilter(common.DefaultDatabase, common.AVAILABLEZONE, availableZone, filter, true)
	if err != nil {
		return nil, err
	}
	return availableZone, nil
}

func (as *AvailableZoneService) Create(reqAvailableZone *system.AvailableZone) (core.IObject, error) {
	if _, err := as.GetByName(reqAvailableZone.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("availableZone exists")
	}

	if reqAvailableZone.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	reqAvailableZone.Kind = system.AvailableZoneKind
	reqAvailableZone.GenerateVersion()

	_, err := as.IService.Create(common.DefaultDatabase, common.AVAILABLEZONE, reqAvailableZone)
	if err != nil {
		return nil, err
	}
	return reqAvailableZone, nil
}

func (as *AvailableZoneService) Update(name string, reqAvailableZone *system.AvailableZone) (core.IObject, bool, error) {
	availableZone, err := as.GetByName(name)
	if err != nil {
		return nil, false, err
	}

	_, update, err := as.IService.Apply(common.DefaultDatabase, common.AVAILABLEZONE, availableZone.Name, availableZone, false)
	if err != nil {
		return nil, false, err
	}

	return availableZone, update, nil
}

func (as *AvailableZoneService) Delete(name string) (core.IObject, error) {
	availableZone, err := as.GetByName(name)
	if err != nil {
		return nil, err
	}
	err = as.DeleteObject(
		common.DefaultDatabase,
		common.AVAILABLEZONE,
		availableZone.Name,
		availableZone,
		true,
	)

	return availableZone, err
}
