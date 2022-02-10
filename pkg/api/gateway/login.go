package gateway

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/core"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	ucfg "github.com/ddx2x/oilmont/pkg/resource/userconfig"
	"github.com/ddx2x/oilmont/pkg/utils/token"
	"golang.org/x/crypto/bcrypt"
)

type userLoginForm struct {
	Name     string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type userCfgF func(tenant string, cfg *ucfg.Config) error

func (gw *Gateway) injectCfg(tenant string, cfg *ucfg.Config) error {
	for _, f := range []userCfgF{gw.allowedWorkspace, gw.allowedResourceOp, gw.allowedProviders, gw.allowedMenus, gw.allowedBiz} {
		if err := f(tenant, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (gw *Gateway) OwnerBiz(tenant, userName string, cfg *ucfg.Config) error {
	//TODO 如果不是admin,且不是租户属主则不执行
	return nil
}

func (gw *Gateway) userCfg(name, tenant, token string, roleType iam.AccountType, isTenantOwner bool) (*ucfg.Config, error) {
	if roleType == iam.AccountTypeAdmin {
		tenant = common.DefaultDatabase
	}
	cfg := &ucfg.Config{
		UserName:      name,
		Tenant:        tenant,
		Token:         token,
		RoleType:      roleType,
		IsTenantOwner: isTenantOwner,
	}

	if !isTenantOwner && roleType == iam.AccountTypeAdmin {
		err := gw.OwnerBiz(tenant, name, cfg)
		if err != nil {
			return nil, err
		}
	}
	if err := gw.injectCfg(tenant, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (gw *Gateway) ThirdCheckAccount(tenant string, user *iam.User) (*iam.Account, error) {
	account, err := gw.GetAccount(tenant, user.Name)
	if err == datasource.NotFound {
		account = &iam.Account{
			Metadata: core.Metadata{
				Name:   user.Name,
				Tenant: tenant,
			},
			Spec: iam.AccountSpec{
				AccountType: iam.AccountTypeTenant,
			},
		}
		_, createErr := gw.stage.Create(tenant, common.ACCOUNT, account)
		if createErr != nil {
			return nil, createErr
		}
	}
	return account, err
}

func (gw *Gateway) ThirdCheckUser(tenant, name, email, enName string) (*iam.User, error) {
	user := &iam.User{
		Metadata: core.Metadata{
			Name:   name,
			Tenant: tenant,
		},
		Spec: iam.UserSpec{
			Email:  email,
			EnName: enName,
			CnName: name,
		},
	}
	err := gw.stage.Get(tenant, common.USER, name, &user, true)
	if err == datasource.NotFound {
		_, createErr := gw.stage.Create(tenant, common.USER, user)
		if createErr != nil {
			return nil, createErr
		}
	}
	return user, err
}

func (gw *Gateway) ThirdLogin(code string) (*ucfg.Config, error) {
	loginType := "feishu" // TODO 需要做第三方类型判断
	data, err := gw.third.GetAccountAccessToken(code)
	if err != nil {
		return nil, err
	}

	userName := strings.Split(data.Email, "@")[0]
	tenantFilter := map[string]interface{}{
		"spec.type": loginType,
		"spec.key":  data.TenantKey,
	}
	tenant, isTenantOwner, err := gw.isTenantOwner("", userName, tenantFilter)
	if err != nil {
		return nil, err
	}

	user, err := gw.ThirdCheckUser(tenant.GetName(), userName, data.Email, data.EnName)
	if err != nil {
		return nil, err
	}

	if !isTenantOwner {
		_, err = gw.ThirdCheckAccount(tenant.GetName(), user)
		if err != nil {
			return nil, err
		}
	}

	userToken, err := token.Encode(tenant.GetName(), userName, common.EXPIRETIME)
	if err != nil {
		return nil, err
	}

	cfg, err := gw.userCfg(userName, tenant.GetName(), userToken, iam.AccountTypeTenant, isTenantOwner)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (gw *Gateway) GetAccount(tenant, name string) (*iam.Account, error) {
	account := &iam.Account{}
	err := gw.stage.Get(tenant, common.ACCOUNT, name, &account, true)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (gw *Gateway) isTenantOwner(tenantName, accountName string, filter map[string]interface{}) (*system.Tenant, bool, error) {
	tenant := &system.Tenant{}
	if filter == nil {
		filter = map[string]interface{}{}
	}
	if tenantName != "" {
		filter[common.FilterName] = tenantName
	}
	err := gw.stage.GetByFilter(common.DefaultDatabase, common.TENANT, &tenant, filter, true)
	if err != nil {
		return nil, false, err
	}
	return tenant, tenant.Spec.Owner == accountName, nil
}

func (gw *Gateway) Login(g *gin.Context) {
	code := g.DefaultQuery("code", "")
	isThirdLogin := code != ""

	if isThirdLogin {
		cfg, err := gw.ThirdLogin(code)
		if err != nil {
			fmt.Println(err)
			api.LoginError(g)
			return
		}
		g.JSON(http.StatusOK, cfg)
		return
	}

	form := &userLoginForm{}
	err := g.ShouldBindJSON(form)
	if err != nil {
		fmt.Println(err)
		api.LoginError(g)
		return
	}

	account, err := gw.GetAccount(common.DefaultDatabase, form.Name)
	if err != nil {
		fmt.Println(err)
		api.LoginError(g)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Spec.Password), []byte(form.Password))
	if err != nil {
		fmt.Println(err)
		api.LoginError(g)
		return
	}

	userToken, err := token.Encode(common.DefaultDatabase, form.Name, common.EXPIRETIME)
	if err != nil {
		api.LoginError(g)
		return
	}

	cfg, err := gw.userCfg(account.GetName(), account.GetTenant(), userToken, account.Spec.AccountType, false)
	if err != nil {
		fmt.Println(err)
		api.LoginError(g)
		return
	}

	g.JSON(http.StatusOK, cfg)
}
