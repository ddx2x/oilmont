package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type TenantService struct {
	service.IService
}

func NewTenant(i service.IService) *TenantService {
	return &TenantService{i}
}

func (as *TenantService) List(name string) (*system.TenantList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}

	data := make([]system.Tenant, 0)
	err := as.IService.ListToObject(common.DefaultDatabase, common.TENANT, filter, &data, true)
	if err != nil {
		return nil, err
	}

	tenantList := &system.TenantList{Items: data}
	tenantList.GenerateListVersion()

	return tenantList, nil
}

func (as *TenantService) GetByName(name string) (*system.Tenant, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}

	tenant := &system.Tenant{}
	err := as.IService.GetByFilter(common.DefaultDatabase, common.TENANT, tenant, filter, true)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (as *TenantService) Create(reqTenant *system.ReqTenant) (core.IObject, error) {
	if _, err := as.GetByName(reqTenant.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("tenant exists")
	}

	if reqTenant.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	reqTenant.Kind = system.TenantKind
	reqTenant.GenerateVersion()

	if err := as.IService.Get(common.DefaultDatabase, common.TENANT, reqTenant.Spec.Owner, system.Tenant{}, true); err != datasource.NotFound {
		return nil, fmt.Errorf("tenant exists")
	}

	_, err := as.IService.Create(common.DefaultDatabase, common.TENANT, reqTenant)
	if err != nil {
		return nil, err
	}

	return reqTenant, nil
}

func (as *TenantService) Update(name string, reqTenant *system.Tenant, path string) (core.IObject, bool, error) {
	new, update, err := as.IService.Apply(common.DefaultDatabase, common.TENANT, name, reqTenant, false, path)
	if err != nil {
		return nil, false, err
	}

	return new, update, nil
}

func (as *TenantService) Delete(name string) (core.IObject, error) {
	tenant, err := as.GetByName(name)
	if err != nil {
		return nil, err
	}
	err = as.DeleteObject(common.DefaultDatabase, common.TENANT, tenant.Name, tenant, true)
	return tenant, err
}
