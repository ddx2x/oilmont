package gateway

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	ucfg "github.com/ddx2x/oilmont/pkg/resource/userconfig"
	"github.com/ddx2x/oilmont/pkg/utils/obj"
)

func (gw *Gateway) getDatabase() ([]string, error) {
	// 每个租户的数据都放在其自己的数据库下，通过租户获取所有数据库
	rawData, err := gw.stage.List(common.DefaultDatabase, common.TENANT, "", true)
	if err != nil {
		return nil, err
	}
	tenants := make([]system.Tenant, 0)

	err = obj.UnstructuredObjectToInstanceObj(&rawData, &tenants)
	if err != nil {
		return nil, err
	}

	databases := make([]string, 0)
	for _, tenant := range tenants {
		databases = append(databases, tenant.GetName())
	}
	// 除租户的数据库外，还有默认的数据库
	databases = append(databases, common.DefaultDatabase)
	return databases, nil
}

func (gw *Gateway) allowedBiz(tenant string, cfg *ucfg.Config) error {
	dbList, err := gw.getDatabase()
	if err != nil {
		return err
	}

	switch cfg.RoleType {
	case iam.AccountTypeAdmin:
		for _, db := range dbList {
			bizGroups := make([]iam.BusinessGroup, 0)
			if err := gw.stage.ListToObject(db, common.BUSINESSGROUP, map[string]interface{}{}, &bizGroups, true); err != nil {
				return err
			}
			for _, group := range bizGroups {
				cfg.BizGroups = append(cfg.BizGroups, group.GetName())
			}
		}
		return nil
	case iam.AccountTypeTenant:
		// tenant owner handle
		if _, isOwner, _ := gw.isTenantOwner(cfg.Tenant, cfg.UserName, nil); isOwner {
			bizGroups := make([]iam.BusinessGroup, 0)
			if err := gw.stage.ListToObject(cfg.Tenant, common.BUSINESSGROUP, map[string]interface{}{}, &bizGroups, true); err != nil {
				return err
			}
			for _, group := range bizGroups {
				cfg.BizGroups = append(cfg.BizGroups, group.GetName())
			}
			return nil
		}

		// Normal user handle
		accountFilter := map[string]interface{}{
			"metadata.name": cfg.UserName,
		}
		account := &iam.Account{}
		if err := gw.stage.GetByFilter(cfg.Tenant, common.ACCOUNT, account, accountFilter, true); err != nil {
			return err
		}
		for biz, _ := range account.Spec.BusinessGroupRole {
			cfg.BizGroups = append(cfg.BizGroups, biz)
		}

		bizGroups := make([]iam.BusinessGroup, 0)
		bizGroupFilter := map[string]interface{}{"spec.owner": cfg.UserName}
		if err := gw.stage.ListToObject(cfg.Tenant, common.BUSINESSGROUP, bizGroupFilter, &bizGroups, true); err != nil {
			return err
		}
		for _, group := range bizGroups {
			cfg.BizGroups = append(cfg.BizGroups, group.GetName())
		}
		return nil

	}
	return nil
}
