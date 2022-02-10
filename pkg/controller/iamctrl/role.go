package iamctrl

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
	"github.com/ddx2x/oilmont/pkg/resource/system"
)

func (r *RBACController) reconcileRole(role *rbac.Role) error {
	if role.IsDelete == true {
		return r.deleteRoleRelation(role)
	}

	return r.refreshResourceWithRole(role)
}

func (r *RBACController) deleteRoleRelation(role *rbac.Role) error {
	if err := r.deleteRoleWithAccountRelation(role); err != nil {
		return err
	}
	if err := r.deleteRoleWithBusinessGroupRelation(role); err != nil {
		return err
	}
	return nil
}

func (r *RBACController) deleteRoleWithAccountRelation(role *rbac.Role) error {
	accountWithRoleRelation := make([]system.Relation, 0)
	roleKey := fmt.Sprintf("spec.resources.%s", common.ROLE)
	relationFilter := map[string]interface{}{
		"spec.relation_kind": common.ACCOUNTROLE,
		roleKey:              role.GetUUID(),
	}

	if err := r.ListToObject(role.GetTenant(), common.RELATION, relationFilter, &accountWithRoleRelation, true); err != nil {
		return err
	}

	for _, relation := range accountWithRoleRelation {
		if err := r.DeleteByUUID(role.GetTenant(), common.RELATION, relation.GetUUID()); err != nil {
			return err
		}

		account := iam.Account{}
		err := r.GetByMetadataUUID(role.GetTenant(), common.ACCOUNT, relation.Spec.Resources[common.ACCOUNT], &account, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		if _, _, err := r.Apply(role.GetTenant(), common.ACCOUNT, account.GetName(), &account, true); err != nil {
			return err
		}
	}
	return nil
}

func (r *RBACController) deleteRoleWithBusinessGroupRelation(role *rbac.Role) error {
	businessGroupWithRoleRelation := make([]system.Relation, 0)
	roleKey := fmt.Sprintf("spec.resources.%s", common.ROLE)
	relationFilter := map[string]interface{}{
		"spec.relation_kind": common.BUSINESSGROUPROLE,
		roleKey:              role.GetUUID(),
	}

	if err := r.ListToObject(role.GetTenant(), common.RELATION, relationFilter, &businessGroupWithRoleRelation, true); err != nil {
		return err
	}

	for _, relation := range businessGroupWithRoleRelation {
		if err := r.DeleteByUUID(role.GetTenant(), common.RELATION, relation.GetUUID()); err != nil {
			return err
		}

		businessGroup := iam.BusinessGroup{}
		err := r.GetByMetadataUUID(role.GetTenant(), common.BUSINESSGROUP, relation.Spec.Resources[common.BUSINESSGROUP], &businessGroup, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		if _, _, err := r.Apply(role.GetTenant(), common.BUSINESSGROUP, businessGroup.GetName(), &businessGroup, true); err != nil {
			return err
		}
	}
	return nil
}

func (r *RBACController) refreshResourceWithRole(role *rbac.Role) error {
	accountWithRoleRelation := make([]system.Relation, 0)
	roleKey := fmt.Sprintf("spec.resources.%s", common.ROLE)
	relationFilter := map[string]interface{}{
		"spec.relation_kind": common.ACCOUNTROLE,
		roleKey:              role.GetUUID(),
	}

	if err := r.ListToObject(role.GetTenant(), common.RELATION, relationFilter, &accountWithRoleRelation, true); err != nil {
		return err
	}

	for _, relation := range accountWithRoleRelation {
		account := iam.Account{}
		err := r.GetByMetadataUUID(role.GetTenant(), common.ACCOUNT, relation.Spec.Resources[common.ACCOUNT], &account, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		if _, _, err := r.Apply(role.GetTenant(), common.ACCOUNT, account.GetName(), &account, true); err != nil {
			return err
		}
	}

	businessGroupWithRoleRelation := make([]system.Relation, 0)
	relationFilter = map[string]interface{}{
		"spec.relation_kind": common.BUSINESSGROUPROLE,
		roleKey:              role.GetUUID(),
	}

	if err := r.ListToObject(role.GetTenant(), common.RELATION, relationFilter, &businessGroupWithRoleRelation, true); err != nil {
		return err
	}

	for _, relation := range businessGroupWithRoleRelation {
		businessGroup := iam.BusinessGroup{}
		err := r.GetByMetadataUUID(role.GetTenant(), common.BUSINESSGROUP, relation.Spec.Resources[common.BUSINESSGROUP], &businessGroup, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return err
		}
		if _, _, err := r.Apply(role.GetTenant(), common.BUSINESSGROUP, businessGroup.GetName(), &businessGroup, true); err != nil {
			return err
		}
	}
	return nil
}
