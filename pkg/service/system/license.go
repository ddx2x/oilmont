package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type LicenseService struct {
	service.IService
}

func NewLicenseService(i service.IService) *LicenseService {
	return &LicenseService{i}
}

func (ls *LicenseService) List(name, workspace string) (*system.LicenseList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	data := make([]system.License, 0)
	err := ls.IService.ListToObject(common.DefaultDatabase, common.LICENSE, filter, &data, true)
	if err != nil {
		return nil, err
	}

	licenseList := &system.LicenseList{Items: data}
	licenseList.GenerateListVersion()

	return licenseList, nil
}

func (ls *LicenseService) GetByName(workspace, name string) (*system.License, error) {
	filter := map[string]interface{}{
		common.FilterName:      name,
		common.FilterWorkspace: workspace,
	}

	license := &system.License{}
	err := ls.IService.GetByFilter(common.DefaultDatabase, common.LICENSE, license, filter, true)
	if err != nil {
		return nil, err
	}
	return license, nil
}

func (ls *LicenseService) Create(reqLicense *system.License) (core.IObject, error) {
	if _, err := ls.GetByName(reqLicense.Workspace, reqLicense.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("license exists")
	}

	if reqLicense.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	reqLicense.Kind = system.LicenseKind
	reqLicense.GenerateVersion()

	_, err := ls.IService.Create(common.DefaultDatabase, common.LICENSE, reqLicense)
	if err != nil {
		return nil, err
	}
	return reqLicense, nil
}

func (ls *LicenseService) Update(workspace, name string, reqLicense *system.License) (core.IObject, bool, error) {
	license, err := ls.GetByName(workspace, name)
	if err != nil {
		return nil, false, err
	}

	license.Spec.Vendor = reqLicense.Spec.Vendor
	license.Spec.Region = reqLicense.Spec.Region
	license.Spec.AvailableZone = reqLicense.Spec.AvailableZone
	license.Spec.SshType = reqLicense.Spec.SshType
	license.Spec.Key = reqLicense.Spec.Key

	_, update, err := ls.IService.Apply(common.DefaultDatabase, common.LICENSE, license.Name, license, false)
	if err != nil {
		return nil, false, err
	}

	return license, update, nil
}

func (ls *LicenseService) Delete(workspace, name string) (core.IObject, error) {
	license, err := ls.GetByName(workspace, name)
	if err != nil {
		return nil, err
	}

	license.Delete()
	_, _, err = ls.Apply(common.DefaultDatabase, common.LICENSE, license.Name, license, true)
	return license, err
}
