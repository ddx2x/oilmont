package iam

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/service"
)

type UserService struct {
	service.IService
}

func NewUser(i service.IService) *UserService {
	return &UserService{i}
}

func (as *UserService) List(tenant, workspace string) (*iam.UserList, error) {
	filter := map[string]interface{}{}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}
	data := make([]iam.User, 0)
	err := as.IService.ListToObject(tenant, common.USER, filter, &data, true)
	if err != nil {
		return nil, err
	}

	resultList := &iam.UserList{Items: data}
	resultList.GenerateListVersion()

	return resultList, nil
}

func (as *UserService) GetByName(tenant, name string) (*iam.User, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}

	result := &iam.User{}
	err := as.IService.GetByFilter(tenant, common.USER, result, filter, true)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (as *UserService) Update(tenant string, reqUser *iam.User) (core.IObject, bool, error) {
	obj, update, err := as.IService.Apply(tenant, common.USER, reqUser.Name, reqUser, false, "spec.account")
	if err != nil {
		return nil, false, err
	}

	return obj, update, nil
}

func (as *UserService) Delete(tenant, name string) (core.IObject, error) {
	obj, err := as.GetByName(tenant, name)
	if err != nil {
		return nil, err
	}
	err = as.DeleteObject(obj.GetTenant(), common.USER, obj.GetName(), obj, true)
	return obj, err
}
