package iam

import (
	"fmt"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/service"
)

type BusinessGroupService struct {
	service.IService
}

func NewBusinessGroup(i service.IService) *BusinessGroupService {
	return &BusinessGroupService{i}
}

func (bgs *BusinessGroupService) List(tenant, workspace string) (*iam.BusinessGroupList, error) {
	filter := map[string]interface{}{}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	bg := make([]iam.BusinessGroup, 0)

	err := bgs.IService.ListToObject(tenant, common.BUSINESSGROUP, filter, &bg, true)
	if err != nil {
		return nil, err
	}

	bgList := &iam.BusinessGroupList{Items: bg}
	bgList.GenerateListVersion()

	return bgList, nil
}

func (bgs *BusinessGroupService) GetByName(tenant, workspace, name string) (*iam.BusinessGroup, error) {
	filter := map[string]interface{}{}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}
	if name != "" {
		filter[common.FilterName] = name
	}

	businessGroup := &iam.BusinessGroup{}
	if err := bgs.IService.GetByFilter(tenant, common.BUSINESSGROUP, businessGroup, filter, true); err != nil {
		return nil, err
	}
	return businessGroup, nil
}

func (bgs *BusinessGroupService) Create(reqBusinessGroup *iam.BusinessGroup) (core.IObject, error) {
	if reqBusinessGroup.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	if _, err := bgs.GetByName(reqBusinessGroup.GetTenant(), reqBusinessGroup.GetWorkspace(), reqBusinessGroup.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("business-group exists")
	}

	reqBusinessGroup.Kind = iam.BusinessGroupKind
	reqBusinessGroup.GenerateVersion()

	_, err := bgs.IService.Create(reqBusinessGroup.GetTenant(), common.BUSINESSGROUP, reqBusinessGroup)
	if err != nil {
		return nil, err
	}
	return reqBusinessGroup, nil
}

func (bgs *BusinessGroupService) Update(tenant, workspace, name string, reqBusinessGroup *iam.BusinessGroup, paths ...string) (core.IObject, bool, error) {
	obj, update, err := bgs.IService.Apply(tenant, common.BUSINESSGROUP, name, reqBusinessGroup, false, paths...)
	if err != nil {
		return nil, false, err
	}

	return obj, update, nil
}

func (bgs *BusinessGroupService) Delete(tenant, workspace, name string) (core.IObject, error) {
	object, err := bgs.GetByName(tenant, workspace, name)
	if err != nil {
		return nil, err
	}
	if err = bgs.DeleteObject(tenant, common.BUSINESSGROUP, name, object, true); err != nil {
		return nil, err
	}
	return object, nil
}
