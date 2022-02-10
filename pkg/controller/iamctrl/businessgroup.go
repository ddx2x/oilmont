package iamctrl

import (
	"fmt"
	"github.com/thoas/go-funk"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
	"github.com/ddx2x/oilmont/pkg/resource/system"
)

func (r *RBACController) reconcileBizGroup(group *iam.BusinessGroup) error {
	if group.Spec.Roles == nil {
		group.Spec.Roles = make([]string, 0)
	}
	if group.IsDelete == true {
		// bizGroup 已经删除并且 relation 还存在，需要删除和 bizGroup 关联的 account relation
		return r.deleteBizGroupRelation(group)
	}
	return r.reconcileBizGroupRelation(group)
}

func (r *RBACController) deleteBizGroupRelation(group *iam.BusinessGroup) error {
	if err := r.deleteBizGroupWithAccountRelation(group); err != nil {
		return err
	}

	if err := r.deleteBizGroupWithRoleRelation(group, []string{}); err != nil {
		return err
	}

	if err := r.deleteBizGroupWorkspace(group); err != nil {
		return err
	}

	return nil
}

// 业务组删除时，同时删除 workspace
func (r *RBACController) deleteBizGroupWorkspace(group *iam.BusinessGroup) error {
	workspace := system.Workspace{}
	workspaceFilter := map[string]interface{}{
		common.FilterName:      group.GetName(),
		common.FilterWorkspace: group.GetTenant(),
	}

	err := r.GetByFilter(common.DefaultDatabase, common.WORKSPACE, &workspace, workspaceFilter, true)
	if err == datasource.NotFound {
		return nil
	}
	if err != nil {
		return err
	}

	return r.Delete(common.DefaultDatabase, common.WORKSPACE, workspace.GetName(), workspace.GetWorkspace())
}

func (r *RBACController) deleteBizGroupWithAccountRelation(group *iam.BusinessGroup) error {
	AccountWithBizGroupRelation := make([]system.Relation, 0)
	bizGroupKey := fmt.Sprintf("spec.resources.%s", common.BUSINESSGROUP)
	relationFilter := map[string]interface{}{
		"spec.relation_kind": common.ACCOUNTBUSINESSGROUP,
		bizGroupKey:          group.GetUUID(),
	}

	if err := r.ListToObject(group.GetTenant(), common.RELATION, relationFilter, &AccountWithBizGroupRelation, true); err != nil {
		return err
	}

	for _, relation := range AccountWithBizGroupRelation {
		if err := r.DeleteByUUID(group.GetTenant(), common.RELATION, relation.GetUUID()); err != nil {
			return err
		}

		account := iam.Account{}
		err := r.GetByMetadataUUID(group.GetTenant(), common.ACCOUNT, relation.Spec.Resources[common.ACCOUNT], &account, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		if _, _, err := r.Apply(account.GetTenant(), common.ACCOUNT, account.GetName(), &account, true); err != nil {
			return err
		}
	}
	return nil
}

func (r *RBACController) deleteBizGroupWithRoleRelation(group *iam.BusinessGroup, oldRole []string) error {
	bizGroupWithRoleRelation := make([]system.Relation, 0)
	bizGroupKey := fmt.Sprintf("spec.resources.%s", common.BUSINESSGROUP)
	relationFilter := map[string]interface{}{
		"spec.relation_kind": common.BUSINESSGROUPROLE,
		bizGroupKey:          group.GetUUID(),
	}

	if err := r.ListToObject(group.GetTenant(), common.RELATION, relationFilter, &bizGroupWithRoleRelation, true); err != nil {
		return err
	}

	for _, relation := range bizGroupWithRoleRelation {
		if funk.ContainsString(oldRole, relation.Spec.Resources[common.ROLE]) {
			continue
		}

		if err := r.DeleteByUUID(group.GetTenant(), common.RELATION, relation.GetUUID()); err != nil {
			return err
		}
	}
	return nil
}

func (r *RBACController) reconcileBizGroupRelation(group *iam.BusinessGroup) error {
	if err := r.reconcileBizGroupWithRoleRelation(group); err != nil {
		return err
	}
	if err := r.refreshResourceWithBizGroup(group); err != nil {
		return err
	}
	if err := r.checkAndCreateWorkspace(group); err != nil {
		return err
	}
	return nil
}

func (r *RBACController) checkAndCreateWorkspace(group *iam.BusinessGroup) error {
	workspace := system.Workspace{}
	workspaceFilter := map[string]interface{}{
		common.FilterName:      group.GetName(),
		common.FilterWorkspace: group.GetTenant(),
	}

	err := r.GetByFilter(common.DefaultDatabase, common.WORKSPACE, &workspace, workspaceFilter, true)
	if err == datasource.NotFound {
		workspace.Workspace = group.GetTenant()
		workspace.Name = group.GetName()
		workspace.Tenant = group.GetTenant()
		if _, createErr := r.Create(common.DefaultDatabase, common.WORKSPACE, &workspace); createErr != nil {
			return createErr
		}
		return nil
	}
	return err
}

func (r *RBACController) reconcileBizGroupWithRoleRelation(group *iam.BusinessGroup) error {
	roles := make([]rbac.Role, 0)

	for _, roleName := range group.Spec.Roles {
		role := rbac.Role{}
		err := r.Get(group.GetTenant(), common.ROLE, roleName, &role, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		roles = append(roles, role)
	}

	group.Spec.Roles = make([]string, 0)
	roleUUID := make([]string, 0)
	for _, role := range roles {
		relation := system.Relation{
			Metadata: core.Metadata{
				Name: fmt.Sprintf("%s-%s", group.GetUUID(), role.GetUUID()),
			},
			Spec: system.RelationSpec{
				RelationKind: common.BUSINESSGROUPROLE,
				Resources: map[string]string{
					common.BUSINESSGROUP: group.GetUUID(),
					common.ROLE:          role.GetUUID(),
				},
			},
		}
		group.Spec.Roles = append(group.Spec.Roles, role.GetName())
		roleUUID = append(roleUUID, role.GetUUID())
		if _, _, err := r.Apply(group.GetTenant(), common.RELATION, relation.GetName(), &relation, false); err != nil {
			return err
		}
	}

	// 增加的 role 添加了绑定, 可能存在去除的 role, 需要删除绑定
	if err := r.deleteBizGroupWithRoleRelation(group, roleUUID); err != nil {
		return err
	}
	_, _, err := r.Apply(group.GetTenant(), common.BUSINESSGROUP, group.GetName(), group, false)
	return err
}

func (r *RBACController) refreshResourceWithBizGroup(group *iam.BusinessGroup) error {
	accountWithBizGroupRelation := make([]system.Relation, 0)
	bizGroupKey := fmt.Sprintf("spec.resources.%s", common.BUSINESSGROUP)
	relationFilter := map[string]interface{}{
		"spec.relation_kind": common.ACCOUNTBUSINESSGROUP,
		bizGroupKey:          group.GetUUID(),
	}

	if err := r.ListToObject(group.GetTenant(), common.RELATION, relationFilter, &accountWithBizGroupRelation, true); err != nil {
		return err
	}

	for _, relation := range accountWithBizGroupRelation {
		account := iam.Account{}
		err := r.GetByMetadataUUID(group.GetTenant(), common.ACCOUNT, relation.Spec.Resources[common.ACCOUNT], &account, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		if _, _, err := r.Apply(group.GetTenant(), common.ACCOUNT, account.GetName(), &account, true); err != nil {
			return err
		}
	}
	return nil
}
