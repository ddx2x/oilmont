package iamctrl

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/system"
)

func (r *RBACController) reconcileWorkspace(workspace *system.Workspace) error {
	if workspace.IsDelete == true {
		return r.deleteWorkspaceRelation(workspace)
	}
	return r.checkAndCreateBizGroup(workspace)
}

// workspace 删除的同时，需要删除业务组
func (r *RBACController) deleteWorkspaceRelation(workspace *system.Workspace) error {
	bizGroup := iam.BusinessGroup{}
	bizGroupFilter := map[string]interface{}{
		common.FilterName: workspace.GetName(),
	}

	err := r.GetByFilter(workspace.GetWorkspace(), common.BUSINESSGROUP, &bizGroup, bizGroupFilter, true)
	if err == datasource.NotFound {
		return nil
	}
	if err != nil {
		return nil
	}

	return r.Delete(workspace.GetWorkspace(), common.BUSINESSGROUP, bizGroup.GetName(), bizGroup.GetWorkspace())

}

func (r *RBACController) checkAndCreateBizGroup(workspace *system.Workspace) error {
	bizGroup := iam.BusinessGroup{}
	bizGroupFilter := map[string]interface{}{
		common.FilterName: workspace.GetName(),
	}

	err := r.GetByFilter(workspace.GetWorkspace(), common.BUSINESSGROUP, &bizGroup, bizGroupFilter, true)
	if err == datasource.NotFound {
		bizGroup.Name = workspace.GetName()
		bizGroup.Workspace = workspace.GetWorkspace()
		bizGroup.Tenant = workspace.GetWorkspace()
		if _, createErr := r.Create(workspace.GetWorkspace(), common.BUSINESSGROUP, &bizGroup); createErr != nil {
			return createErr
		}
	}
	return nil
}
