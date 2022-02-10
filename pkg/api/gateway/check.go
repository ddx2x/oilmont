package gateway

import (
	"context"
	"fmt"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"go.uber.org/multierr"
	"net/http"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/micro/gateway"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/service"
	"github.com/ddx2x/oilmont/pkg/utils/obj"
	"github.com/ddx2x/oilmont/pkg/utils/uri"
)

var _ InterceptMiddleware = &permission{}

var operationMap = map[string]string{}

var uriParser = uri.NewURIParser()

type permission struct {
	service.IService
}

func newPermission(stage datasource.IStorage) *permission {
	return &permission{
		service.NewBaseService(stage, nil),
	}
}

func (p *permission) Intercept() gateway.Intercept {
	return func(w http.ResponseWriter, r *http.Request) gateway.InterceptType {
		ctx := context.TODO()
		flog := log.G(ctx).WithField("intercept", "perm check")
		userName := r.Header.Get(common.HttpRequestUserHeaderKey)
		tenant := r.Header.Get(common.HttpRequestUserHeaderBELONGTENANT)

		if userName == "" {
			return gateway.NotAuthorized
		}

		// check admin
		if userName == "admin" {
			return gateway.Redirect
		}

		// check tenant owner
		next, pass, err := p.tenantOwnerCheck(tenant, userName)
		if err != nil {
			return gateway.NotAuthorized
		}

		if pass {
			return gateway.Redirect
		}

		if !next {
			return gateway.NotAuthorized
		}

		account := &iam.Account{}
		accountFilter := map[string]interface{}{"metadata.name": userName}
		if err := p.GetByFilter(tenant, common.ACCOUNT, account, accountFilter, true); err != nil {
			return gateway.NotAuthorized
		}

		// TODO: userIdentification
		if len(operationMap) == 0 {
			if err := p.reconcileOperationMap(); err != nil {
				return gateway.NotAuthorized
			}
		}

		uriOp, err := uriParser.ParseOp(r.Method, fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery), operationMap)
		if err != nil {
			flog.Infof("uri parse error: %s", err)
			return gateway.NotAuthorized
		}

		//check ?
		check, err := p.check(account, uriOp)
		if err != nil || !check {
			flog.Infof("check perm error: %s", err)
			return gateway.Redirect
		}

		_ = uriOp

		return gateway.Redirect
	}
}

func (p *permission) checkURIsAndReorganize(tenant, userName string, uris []*uri.URI) ([]*uri.URI, error) {
	var err error
	var res = make([]*uri.URI, 0)
	account := &iam.Account{}
	accountFilter := map[string]interface{}{"metadata.name": userName}

	if err := p.GetByFilter(tenant, common.ACCOUNT, account, accountFilter, true); err != nil {
		return nil, fmt.Errorf("account not exists")
	}

	for _, uri := range uris {
		pass, checkErr := p.check(account, uri)
		if checkErr != nil {
			checkErr = multierr.Append(err, checkErr)
		}
		if !pass || checkErr != nil {
			continue
		}
		res = append(res, uri)
	}
	return res, nil
}

// TODO: 如果是租户owner，直接跳过
func (p *permission) tenantOwnerCheck(tenantName, userName string) (next bool, pass bool, err error) {
	tenant := &system.Tenant{}

	if err := p.Get(common.DefaultDatabase, common.TENANT, tenantName, tenant, true); err != nil {
		return false, false, err
	}
	if userName == tenant.Spec.Owner {
		return false, true, nil
	}
	return true, false, nil
}

func (p *permission) check(account *iam.Account, u *uri.URI) (bool, error) {
	action := u.Op
	resourceName := u.Resource
	workspace := u.Namespace

	resource := system.Resource{}
	resourceFilter := map[string]interface{}{"spec.resourceName": resourceName}
	if err := p.GetByFilter(common.DefaultDatabase, common.RESOURCE, &resource, resourceFilter, true); err != nil {
		return false, err
	}

	if resource.Spec.Type == system.ResourceTypePublic {
		return true, nil
	}

	ap := rbac.AccountPermission{}
	permissionFilter := map[string]interface{}{"spec.account": account.UUID}
	if err := p.GetByFilter(account.GetTenant(), common.ACCOUNTPERMISSION, &ap, permissionFilter, true); err != nil {
		return false, err
	}

	// 检查是否可以操作workspace
	if _, exist := ap.Spec.Permission[workspace]; !exist {
		return false, nil
	}

	// 检查是否可以操作资源
	if _, exist := ap.Spec.Permission[workspace][resourceName]; !exist {
		return false, nil
	}

	//检查是否有资源的操作权限
	if _, exist := ap.Spec.Permission[workspace][resourceName][action]; exist {
		return true, nil // check 需要返过来，默认return false 检查有权限则通过
	}

	return false, nil
}

func (p *permission) reconcileOperationMap() error {
	rawOperation, err := p.List(common.DefaultDatabase, common.OPERATION, "", true)
	if err != nil {
		return err
	}

	operations := make([]system.Operation, len(rawOperation))
	if err = obj.UnstructuredObjectToInstanceObj(&rawOperation, &operations); err != nil {
		return err
	}

	for _, operation := range operations {
		operationMap[operation.Spec.Method] = operation.Spec.OP
	}

	return nil
}
