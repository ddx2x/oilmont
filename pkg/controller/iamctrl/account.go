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
	"github.com/ddx2x/oilmont/pkg/utils/obj"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

func (r *RBACController) reconcileAccount(account *iam.Account) error {
	//if account.Spec.BusinessGroup == nil {
	//	account.Spec.BusinessGroup = make([]string, 0)
	//}
	if account.IsDelete == true {
		return r.deleteAccountRelation(account)
	}
	return r.reconcileAccountRelation(account)
}

func (r *RBACController) deleteAccountRelation(account *iam.Account) error {
	if err := r.deleteAccountWithRoleRelation(account, []string{}); err != nil {
		return err
	}
	if err := r.deleteAccountWithBizGroupRelation(account, []string{}); err != nil {
		return err
	}

	return nil
}

func (r *RBACController) deleteAccountWithRoleRelation(account *iam.Account, oldRole []string) error {
	accountWithRoleRelation := make([]system.Relation, 0)
	accountKey := fmt.Sprintf("spec.resources.%s", common.ACCOUNT)
	relationFilter := map[string]interface{}{
		accountKey:           account.GetUUID(),
		"spec.relation_kind": common.ACCOUNTROLE,
	}

	if err := r.ListToObject(account.GetTenant(), common.RELATION, relationFilter, &accountWithRoleRelation, true); err != nil {
		return err
	}

	for _, relation := range accountWithRoleRelation {
		if funk.ContainsString(oldRole, relation.Spec.Resources[common.ROLE]) {
			continue
		}
		if err := r.DeleteByUUID(account.GetTenant(), common.RELATION, relation.GetUUID()); err != nil {
			return err
		}
	}
	return nil
}

func (r *RBACController) deleteAccountWithBizGroupRelation(account *iam.Account, oldBizGroup []string) error {
	accountWithBizGroupRelation := make([]system.Relation, 0)
	bizGroupKey := fmt.Sprintf("spec.resources.%s", common.BUSINESSGROUP)
	relationFilter := map[string]interface{}{
		bizGroupKey:          account.GetUUID(),
		"spec.relation_kind": common.ACCOUNTBUSINESSGROUP,
	}

	if err := r.ListToObject(account.GetTenant(), common.RELATION, relationFilter, &accountWithBizGroupRelation, true); err != nil {
		return err
	}

	for _, relation := range accountWithBizGroupRelation {
		if funk.ContainsString(oldBizGroup, relation.Spec.Resources[common.BUSINESSGROUP]) {
			continue
		}
		if err := r.DeleteByUUID(account.GetTenant(), common.RELATION, relation.GetUUID()); err != nil {
			return err
		}
	}
	return nil
}

func (r *RBACController) reconcileAccountRelation(account *iam.Account) error {
	if err := r.reconcileAccountWithRoleRelation(account); err != nil {
		return err
	}
	if err := r.reconcileAccountWithBizGroupRelation(account); err != nil {
		return err
	}
	return r.reconcileAccountWithPermission(account)
}

func (r *RBACController) AccountRoleInBizGroupHandle(account *iam.Account, bizGroup string) ([]string, []string, error) {
	roleUUID := make([]string, 0)
	roles := make([]string, 0)
	for _, roleName := range account.Spec.BusinessGroupRole[bizGroup] {
		role := rbac.Role{}
		err := r.Get(account.GetTenant(), common.ROLE, roleName, &role, true)
		if err == datasource.NotFound {
			continue
		}
		if err != nil {
			return nil, nil, err
		}

		relation := system.Relation{
			Metadata: core.Metadata{
				Name: fmt.Sprintf("%s-%s", account.GetUUID(), role.GetUUID()),
			},
			Spec: system.RelationSpec{
				RelationKind: common.ACCOUNTROLE,
				Resources: map[string]string{
					common.ACCOUNT: account.GetUUID(),
					common.ROLE:    role.GetUUID(),
				},
			},
		}
		roleUUID = append(roleUUID, role.GetUUID())
		roles = append(roles, role.GetName())
		if _, _, err := r.Apply(account.GetTenant(), common.RELATION, relation.GetName(), &relation, false); err != nil {
			return nil, nil, err
		}
	}
	return roleUUID, roles, nil
}

func (r *RBACController) reconcileAccountWithRoleRelation(account *iam.Account) error {
	roleUUID := make([]string, 0)
	// 处理 account 拥有的每个业务组的 role
	for bizGroup, roles := range account.Spec.BusinessGroupRole {
		if len(roles) == 0 {
			continue
		}
		bizGroupRoleUUID, rolesName, err := r.AccountRoleInBizGroupHandle(account, bizGroup)
		if err != nil {
			return err
		}
		roleUUID = append(roleUUID, bizGroupRoleUUID...)
		roles = rolesName
	}

	// 增加的 role 添加了绑定, 可能存在去除的 role, 需要删除绑定
	if err := r.deleteAccountWithRoleRelation(account, roleUUID); err != nil {
		return err
	}
	_, _, err := r.Apply(account.GetTenant(), common.ACCOUNT, account.GetName(), account, false)
	return err
}

func (r *RBACController) reconcileAccountWithBizGroupRelation(account *iam.Account) error {
	bizGroups := make([]iam.BusinessGroup, 0)

	for bizGroupName, _ := range account.Spec.BusinessGroupRole {
		bizGroup := iam.BusinessGroup{}
		err := r.Get(account.GetTenant(), common.BUSINESSGROUP, bizGroupName, &bizGroup, true)
		if err == datasource.NotFound {
			delete(account.Spec.BusinessGroupRole, bizGroupName)
			continue
		}
		if err != nil {
			return err
		}
		bizGroups = append(bizGroups, bizGroup)
	}

	bizGroupUUID := make([]string, 0)
	for _, bizGroup := range bizGroups {
		relation := system.Relation{
			Metadata: core.Metadata{
				Name: fmt.Sprintf("%s-%s", account.GetUUID(), bizGroup.GetUUID()),
			},
			Spec: system.RelationSpec{
				RelationKind: common.ACCOUNTBUSINESSGROUP,
				Resources: map[string]string{
					common.ACCOUNT:       account.GetUUID(),
					common.BUSINESSGROUP: bizGroup.GetUUID(),
				},
			},
		}
		bizGroupUUID = append(bizGroupUUID, bizGroup.GetUUID())
		if _, _, err := r.Apply(account.GetTenant(), common.RELATION, relation.GetName(), &relation, false); err != nil {
			return err
		}
	}

	// 增加的 bizGroup 添加了绑定, 可能存在去除的 bizGroup, 需要删除绑定
	if err := r.deleteAccountWithBizGroupRelation(account, bizGroupUUID); err != nil {
		return err
	}
	_, _, err := r.Apply(account.GetTenant(), common.ACCOUNT, account.GetName(), account, false)
	return err
}

func (r *RBACController) RolePermissionHandle(actionMenus map[string]interface{}, permissions map[string]map[string]struct{}, menusName *[]string) error {
	for resource, rawStrings := range actionMenus {
		if reflect.TypeOf(rawStrings).Kind() == reflect.Map {
			if err := r.RolePermissionHandle(rawStrings.(map[string]interface{}), permissions, menusName); err != nil {
				return err
			}
			continue
		}
		resourceObj := &system.Resource{}
		resourceFilter := map[string]interface{}{
			"spec.menu": resource,
		}
		if err := r.GetByFilter(common.DefaultDatabase, common.RESOURCE, resourceObj, resourceFilter, true); err != nil {
			continue
		}

		if _, exist := permissions[resourceObj.Spec.ResourceName]; !exist {
			permissions[resourceObj.Spec.ResourceName] = map[string]struct{}{}
		}

		strings := make([]string, 0)
		if err := obj.UnstructuredObjectToInstanceObj(rawStrings, &strings); err != nil {
			return err
		}

		for _, opName := range strings {
			op := &system.Operation{}
			if err := r.Get(common.DefaultDatabase, common.OPERATION, opName, op, true); err != nil {
				continue
			}
			permissions[resourceObj.Spec.ResourceName][op.Spec.OP] = struct{}{}
		}
		*menusName = append(*menusName, resource)
	}
	return nil
}

func (r *RBACController) getPermissionAndMenuByRole(account *iam.Account, bizGroup *iam.BusinessGroup) (map[string]map[string]struct{}, []string, error) {
	roles := make([]rbac.Role, 0)
	rolesName := account.Spec.BusinessGroupRole[bizGroup.GetName()]
	roleFilter := map[string]interface{}{}

	if bizGroup.Spec.Owner == account.GetName() {
		roleFilter["spec.business"] = bizGroup.GetName()
	} else {
		roleFilter["metadata.name"] = bson.M{"$in": rolesName}
	}

	if err := r.ListToObject(account.GetTenant(), common.ROLE, roleFilter, &roles, true); err != nil {
		return nil, nil, err
	}

	if len(roles) == 0 {
		return nil, nil, nil
	}

	permissions := make(map[string]map[string]struct{}) //  {"resource-1":{"op-1":{},"op-2":{}}}
	menusName := make([]string, 0)

	for _, role := range roles {
		for productMenus, rawActionMenus := range role.Spec.Permission {
			if err := r.RolePermissionHandle(rawActionMenus.(map[string]interface{}), permissions, &menusName); err != nil {
				return nil, nil, err
			}
			menusName = append(menusName, productMenus)
		}
	}
	menusName = funk.UniqString(menusName)
	return permissions, menusName, nil
}

func (r *RBACController) addMenuFromResource(menuName string, menus *[]string) error {
	menu := system.Menu{}
	menuFilter := map[string]interface{}{common.FilterName: menuName}
	if err := r.GetByFilter(common.DefaultDatabase, common.Menu, &menu, menuFilter, true); err != nil {
		return err
	}
	*menus = append(*menus, menuName)
	if menu.Spec.Parent != "" {
		return r.addMenuFromResource(menu.Spec.Parent, menus)
	}
	return nil
}

func (r *RBACController) reconcileAccountWithPermission(account *iam.Account) error {
	permission := make(rbac.AccountPermissionMap)
	menus := make([]string, 0)
	workspace := make([]string, 0)

	//查询所有 bizGroup 属主
	bizGroups := make([]iam.BusinessGroup, 0)
	bizGroupFilter := map[string]interface{}{
		"spec.owner": account.GetName(),
	}
	if err := r.ListToObject(account.GetTenant(), common.BUSINESSGROUP, bizGroupFilter, &bizGroups, true); err != nil {
		return err
	}

	for _, group := range bizGroups {
		account.Spec.BusinessGroupRole[group.GetName()] = make([]string, 0)
	}

	for bizGroupName, _ := range account.Spec.BusinessGroupRole {
		bizGroup := &iam.BusinessGroup{}
		if err := r.Get(account.GetTenant(), common.BUSINESSGROUP, bizGroupName, bizGroup, true); err != nil {
			return err
		}

		rolePermission, roleMenus, err := r.getPermissionAndMenuByRole(account, bizGroup)
		if err != nil {
			return err
		}
		permission[bizGroupName] = rolePermission

		// 如果是属主，增加属于属主的资源
		if bizGroup.Spec.Owner == account.GetName() {
			resources := make([]system.Resource, 0)
			resourceFilter := map[string]interface{}{
				"spec.type_of_role": system.ResourceTypeOfRoleBizOwner,
			}
			if err := r.ListToObject(common.DefaultDatabase, common.RESOURCE, resourceFilter, &resources, true); err != nil {
				return err
			}

			if permission[bizGroupName] == nil {
				permission[bizGroupName] = make(map[string]map[string]struct{})
			}
			for _, resource := range resources {
				if permission[bizGroupName][resource.Spec.ResourceName] == nil {
					permission[bizGroupName][resource.Spec.ResourceName] = make(map[string]struct{})
				}
				for _, op := range resource.Spec.Ops {
					permission[bizGroupName][resource.Spec.ResourceName][op] = struct{}{}
				}

				//可能只有子目录，需要获取父目录
				if err := r.addMenuFromResource(resource.Spec.Menu, &menus); err != nil {
					return err
				}
			}
		}

		menus = append(menus, roleMenus...)
		workspace = append(workspace, bizGroup.GetName())
	}

	menus = funk.UniqString(menus)

	accountPermission := &rbac.AccountPermission{
		Metadata: core.Metadata{
			Name: account.GetName(),
		},
		Spec: rbac.AccountPermissionSpec{
			Account:    account.GetUUID(),
			Permission: permission,
			Menus:      menus,
			Workspaces: workspace,
		},
	}

	if err := r.Delete(account.GetTenant(), common.ACCOUNTPERMISSION, accountPermission.Name, ""); err != nil {
		return err
	}

	_, _, err := r.Apply(account.GetTenant(), common.ACCOUNTPERMISSION, accountPermission.Name, accountPermission, true)
	return err
}
