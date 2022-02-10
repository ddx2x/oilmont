package gateway

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	ucfg "github.com/ddx2x/oilmont/pkg/resource/userconfig"
	"go.mongodb.org/mongo-driver/bson"
)

func (gw *Gateway) allowedMenusAdminHandle(tenant string, cfg *ucfg.Config) error {
	userMenuTrees := new(ucfg.UserMenuTrees)
	productMenus := make([]*system.Menu, 0)
	filter := map[string]interface{}{"spec.type": "product"}
	err := gw.stage.ListToObject(tenant, common.Menu, filter, &productMenus, true)
	if err != nil {
		return err
	}
	for _, _menu := range productMenus {
		menu := _menu
		um := &ucfg.UserMenu{
			Name:   menu.GetName(),
			Link:   menu.Spec.Link,
			Title:  menu.Spec.Title,
			Icon:   menu.Spec.Icon,
			Parent: !menu.Spec.IsSubMenu,
		}
		userMenuTrees.AddRoot(um)
	}

	actionParentMenus := make([]*system.Menu, 0)
	filter = map[string]interface{}{"spec.level": 2}
	err = gw.stage.ListToObject(tenant, common.Menu, filter, &actionParentMenus, true)
	if err != nil {
		return err
	}

	for _, _menu := range actionParentMenus {
		menu := _menu
		userMenuTrees.AddBranch(
			&ucfg.UserMenu{
				Name:   menu.GetName(),
				Link:   menu.Spec.Link,
				Title:  menu.Spec.Title,
				Icon:   menu.Spec.Icon,
				Parent: !menu.Spec.IsSubMenu,
			},
			menu.Spec.Parent,
		)
	}

	actionMenus := make([]*system.Menu, 0)
	filter = map[string]interface{}{"spec.level": 3}
	err = gw.stage.ListToObject(tenant, common.Menu, filter, &actionMenus, true)
	if err != nil {
		return err
	}

	for _, _menu := range actionMenus {
		menu := _menu
		userMenuTrees.AddLeaf(
			&ucfg.UserMenu{
				Name:   menu.GetName(),
				Link:   menu.Spec.Link,
				Title:  menu.Spec.Title,
				Icon:   menu.Spec.Icon,
				Parent: !menu.Spec.IsSubMenu,
			},
			menu.Spec.Parent,
		)
	}
	cfg.Menus = *userMenuTrees

	return nil
}

func (gw *Gateway) allowedMenusTenantOwnerHandle(tenant string, cfg *ucfg.Config) error {
	menuFilter := make([]string, 0)
	getTenant, err := gw.getTenant(tenant)
	if err != nil {
		return err
	}
	for product, action := range getTenant.Spec.Permission {
		menuFilter = append(menuFilter, product)
		if _, ok := action.(string); ok {
			menuFilter = append(menuFilter, action.(string))
		}
		if _, ok := action.(map[string]interface{}); ok {
			for k, _ := range action.(map[string]interface{}) {
				menuFilter = append(menuFilter, k)
			}
		}
	}
	return gw.injectAllowedMenus(cfg, menuFilter)
}

func (gw *Gateway) allowedMenusTenantHandle(tenant string, cfg *ucfg.Config) error {
	menuFilter := make([]string, 0)
	_, isTenantOwner, _ := gw.isTenantOwner(tenant, cfg.UserName, nil)
	if isTenantOwner {
		return gw.allowedMenusTenantOwnerHandle(tenant, cfg)
	}

	// Normal user handle
	filter := map[string]interface{}{
		"metadata.name": cfg.UserName,
	}
	ap := rbac.AccountPermission{}
	if err := gw.stage.GetByFilter(cfg.Tenant, common.ACCOUNTPERMISSION, &ap, filter, true); err != nil {
		return err
	}
	menuFilter = ap.Spec.Menus
	if err := gw.injectAllowedMenus(cfg, menuFilter); err != nil {
		return err
	}

	return nil
}

func (gw *Gateway) injectAllowedMenus(cfg *ucfg.Config, menuFilter []string) error {
	userMenuTrees := new(ucfg.UserMenuTrees)
	productMenus := make([]*system.Menu, 0)
	err := gw.stage.ListToObject(
		common.DefaultDatabase,
		common.Menu,
		map[string]interface{}{
			"metadata.name": bson.M{"$in": menuFilter},
			"spec.type":     "product",
		}, &productMenus,
		true,
	)
	if err != nil {
		return err
	}

	for _, _menu := range productMenus {
		menu := _menu
		um := &ucfg.UserMenu{
			Name:   menu.GetName(),
			Link:   menu.Spec.Link,
			Title:  menu.Spec.Title,
			Icon:   menu.Spec.Icon,
			Parent: !menu.Spec.IsSubMenu,
		}
		userMenuTrees.AddRoot(um)
	}

	actionParentMenus := make([]*system.Menu, 0)
	if err := gw.stage.ListToObject(
		common.DefaultDatabase,
		common.Menu,
		map[string]interface{}{
			"metadata.name": bson.M{"$in": menuFilter},
			"spec.level":    2,
		}, &actionParentMenus, true,
	); err != nil {
		return err
	}

	for _, _menu := range actionParentMenus {
		menu := _menu
		userMenuTrees.AddBranch(
			&ucfg.UserMenu{
				Name:   menu.GetName(),
				Link:   menu.Spec.Link,
				Title:  menu.Spec.Title,
				Icon:   menu.Spec.Icon,
				Parent: !menu.Spec.IsSubMenu,
			},
			menu.Spec.Parent,
		)
	}

	actionMenus := make([]*system.Menu, 0)
	if err := gw.stage.ListToObject(
		common.DefaultDatabase,
		common.Menu,
		map[string]interface{}{
			"metadata.name": bson.M{"$in": menuFilter},
			"spec.level":    3,
		}, &actionMenus, true,
	); err != nil {
		return err
	}

	for _, _menu := range actionMenus {
		menu := _menu
		userMenuTrees.AddLeaf(
			&ucfg.UserMenu{
				Name:   menu.GetName(),
				Link:   menu.Spec.Link,
				Title:  menu.Spec.Title,
				Icon:   menu.Spec.Icon,
				Parent: !menu.Spec.IsSubMenu,
			},
			menu.Spec.Parent,
		)
	}

	cfg.Menus = *userMenuTrees
	return nil
}

func (gw *Gateway) allowedMenus(tenant string, cfg *ucfg.Config) error {
	if cfg.RoleType == iam.AccountTypeAdmin {
		return gw.allowedMenusAdminHandle(tenant, cfg)
	}
	return gw.allowedMenusTenantHandle(tenant, cfg)
}
