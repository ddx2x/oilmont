package iam

import (
	"fmt"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
	"github.com/ddx2x/oilmont/pkg/service"
)

type RoleService struct {
	service.IService
}

func NewRole(i service.IService) *RoleService {
	return &RoleService{i}
}

func (r *RoleService) List(tenant, workspace, name string) (*rbac.RoleList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	roles := make([]rbac.Role, 0)
	err := r.IService.ListToObject(tenant, common.ROLE, filter, &roles, true)
	if err != nil {
		return nil, err
	}

	roleList := &rbac.RoleList{Items: roles}
	roleList.GenerateListVersion()

	return roleList, nil
}

func (r *RoleService) GetByName(tenant, workspace, name string) (*rbac.Role, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	role := &rbac.Role{}
	if err := r.IService.GetByFilter(tenant, common.ROLE, role, filter, true); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleService) Create(reqRole *rbac.Role) (*rbac.Role, error) {
	if reqRole.Name == "" {
		return nil, fmt.Errorf("role name is empty")
	}

	if _, err := r.GetByName(reqRole.GetTenant(), reqRole.GetWorkspace(), reqRole.Name); err != datasource.NotFound {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("role exists")
	}

	reqRole.Kind = rbac.RoleKind
	reqRole.GenerateVersion()

	_, err := r.IService.Create(reqRole.GetTenant(), common.ROLE, reqRole)
	if err != nil {
		return nil, err
	}

	return reqRole, nil
}

func (r *RoleService) Update(tenant, workspace, name string, reqRole *rbac.Role, paths ...string) (core.IObject, bool, error) {
	reqRole.Workspace = workspace

	newObj, update, err := r.IService.Apply(tenant, common.ROLE, name, reqRole, true, paths...)
	if err != nil {
		return nil, false, err
	}

	return newObj, update, nil
}

func (r *RoleService) Delete(tenant, workspace, name string) (core.IObject, error) {
	object, err := r.GetByName(tenant, workspace, name)
	if err != nil {
		return nil, err
	}
	if err = r.DeleteObject(tenant, common.ROLE, name, object, true); err != nil {
		return nil, err
	}
	return object, err
}
