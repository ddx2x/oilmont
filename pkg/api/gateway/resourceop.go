package gateway

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	ucfg "github.com/ddx2x/oilmont/pkg/resource/userconfig"
	"go.mongodb.org/mongo-driver/bson"
)

func (gw *Gateway) getTenant(name string) (*system.Tenant, error) {
	tenant := &system.Tenant{}
	if err := gw.stage.Get(common.DefaultDatabase, common.TENANT, name, tenant, true); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (gw *Gateway) allowedWorkspace(tenant string, cfg *ucfg.Config) error {
	workspaces := make([]system.Workspace, 0)
	filter := map[string]interface{}{}
	switch cfg.RoleType {
	case iam.AccountTypeAdmin:
	case iam.AccountTypeTenant: // TenantOwner
		if _, isOwner, _ := gw.isTenantOwner(cfg.Tenant, cfg.UserName, nil); isOwner {
			// Tenant owner handle
			filter["metadata.workspace"] = tenant
		} else {
			// Normal user handle
			apFilter := map[string]interface{}{
				"metadata.name": cfg.UserName,
			}
			ap := rbac.AccountPermission{}
			if err := gw.stage.GetByFilter(cfg.Tenant, common.ACCOUNTPERMISSION, &ap, apFilter, true); err != nil {
				return err
			}
			filter["metadata.name"] = bson.M{"$in": ap.Spec.Workspaces}
		}
	}
	err := gw.stage.ListToObject(common.DefaultDatabase, common.WORKSPACE, filter, &workspaces, true)
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		cfg.AllowedWorkspaces = append(cfg.AllowedWorkspaces, workspace.GetName())
	}
	return nil
}

func (gw *Gateway) getTenantPerm(tenantName string) (map[string]interface{}, error) {
	tenant, err := gw.getTenant(tenantName)
	if err != nil {
		return nil, err
	}
	return tenant.Spec.Permission, nil
}

func (gw *Gateway) listResourceByMenu(menuName string) ([]system.Resource, error) {
	var resources []system.Resource
	err := gw.stage.ListToObject(common.DefaultDatabase, common.RESOURCE, map[string]interface{}{"spec.menu": menuName}, &resources, true)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (gw *Gateway) allowedResourceOp(tenant string, cfg *ucfg.Config) error {

	var menus []system.Menu
	// is admin
	if cfg.RoleType == iam.AccountTypeAdmin {
		err := gw.stage.ListToObject(common.DefaultDatabase, common.Menu, nil, &menus, true)
		if err != nil {
			return err
		}

		if cfg.Permission == nil {
			cfg.Permission = make(map[string]interface{})
		}

		for _, menu := range menus {
			menuName := menu.GetName()
			resources, err := gw.listResourceByMenu(menuName)
			if err != nil {
				return err
			}
			for _, resource := range resources {
				for _, op := range resource.Spec.Ops {
					filter := map[string]interface{}{"metadata.name": op}
					operation := &system.Operation{}
					err := gw.stage.GetByFilter(common.DefaultDatabase, common.OPERATION, &operation, filter, true)
					if err != nil {
						if err == datasource.NotFound {
							continue
						}
						return err
					}
					resourceName := resource.GetName()
					var product map[string][]string
					if _, exist := cfg.Permission[menuName]; !exist {
						product = make(map[string][]string, 0)
						cfg.Permission[menuName] = product
					} else {
						product = cfg.Permission[menuName].(map[string][]string)
					}

					var action []string
					if _, exist := product[resourceName]; !exist {
						action = make([]string, 0)
						product[resourceName] = action
					} else {
						action = product[resourceName]
					}
					action = append(action, op)
				}
			}
		}

		return nil
	}

	// is tenant owner
	if _, isTenantOwner, _ := gw.isTenantOwner(tenant, cfg.UserName, nil); isTenantOwner {
		perms, err := gw.getTenantPerm(tenant)
		if err != nil {
			return err
		}
		cfg.Permission = perms
		return nil
	}

	// other
	filter := map[string]interface{}{
		"metadata.name": cfg.UserName,
	}
	accountPermission := rbac.AccountPermission{}
	if err := gw.stage.GetByFilter(cfg.Tenant, common.ACCOUNTPERMISSION, &accountPermission, filter, true); err != nil && err != datasource.NotFound {
		return err
	}

	for resourceEnName, op := range accountPermission.Spec.Permission {
		resourceCnName, ok := gw.cache.GetCache(resourceEnName)
		if !ok {
			resourceFilter := map[string]interface{}{
				"spec.resourceName": resourceEnName,
			}
			resource := system.Resource{}
			if err := gw.stage.GetByFilter(common.DefaultDatabase, common.RESOURCE, &resource, resourceFilter, true); err != nil {
				if err == datasource.NotFound {
					continue
				}
				return err
			}
			resourceCnName = resource.GetName()
			gw.cache.SetCache(resourceEnName, resourceCnName, -1)
		}

		ops := make([]string, 0)
		for opEnName, _ := range op {
			opCnName, ok := gw.cache.GetCache(opEnName)
			if !ok {
				opFilter := map[string]interface{}{
					"spec.op": opEnName,
				}
				opObj := system.Operation{}
				if err := gw.stage.GetByFilter(common.DefaultDatabase, common.OPERATION, &opObj, opFilter, true); err != nil {
					return err
				}
				opCnName = opObj.GetName()
				gw.cache.SetCache(opEnName, opCnName, -1)
			}
			ops = append(ops, opCnName.(string))
		}

		cfg.Permission[resourceCnName.(string)] = ops
	}

	return nil
}
