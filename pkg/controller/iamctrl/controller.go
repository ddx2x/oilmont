package iamctrl

import (
	"context"
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/controller"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/ddx2x/oilmont/pkg/proc"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/resource/rbac"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/utils/obj"
	"github.com/ddx2x/oilmont/pkg/utils/thirdlogin/feishu"
	"time"
)

var _ controller.Controller = &RBACController{}

type RBACController struct {
	datasource.IStorage
	proc *proc.Proc
	flog log.Logger
}

func NewRBACController(store datasource.IStorage) *RBACController {
	flog := log.GetLogger(context.Background()).WithField("controller", "iamctrl")
	server := &RBACController{
		IStorage: store,
		proc:     proc.NewProc(),
		flog:     flog,
	}
	return server
}

func (r *RBACController) Run() error {
	r.proc.Add(r.WatchAccount, r.WatchBizGroup, r.WatchWorkspace, r.WatchRole)
	return <-r.proc.Start()
}

func (r *RBACController) WatchAccount(errC chan<- error) {
	r.flog.Infof("RBACController start watch account")
	flog := r.flog.WithField("thread", "account")

	accountCoder := datasource.GetCoder(string(iam.AccountKind))
	if accountCoder == nil {
		errC <- fmt.Errorf("(%s) %s", iam.AccountKind, "coder not exist")
	}

	accountWatchChan := datasource.NewWatch(accountCoder)
	var version = "0"
	dbList, err := r.getDatabase()
	if err != nil {
		errC <- err
	}
	for _, s := range dbList {
		rawAccounts, err := r.List(s, common.ACCOUNT, "", false)
		if err != nil {
			errC <- err
		}
		for _, item := range rawAccounts {
			account := &iam.Account{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, account); err != nil {
				flog.Infof("unstructured account %s, tenant:%s error %s\n", account.GetName(), account.GetTenant(), err)
			}
			if account.GetResourceVersion() > version {
				version = account.GetResourceVersion()
			}
			flog.Infof("get reconcile account %s\n", account.Name)

			if err := r.reconcileAccount(account); err != nil {
				flog.Infof("reconcile account %s, tenant:%s error %s\n", account.GetName(), account.GetTenant(), err)
			}
		}
	}

	for _, s := range dbList {
		r.Watch(s, common.ACCOUNT, version, accountWatchChan)
	}

	for {
		select {
		case item, ok := <-accountWatchChan.ResultChan():
			if !ok {
				return
			}

			account := &iam.Account{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, account); err != nil {
				flog.Infof("watch account unstruct error %s\n", err)
				continue
			}

			flog.Infof("watch reconcile account %s\n", account.Name)
			if err := r.reconcileAccount(account); err != nil {
				flog.Infof("watch account reconcile error %s\n", err)
			}
		}
	}
}

func (r *RBACController) WatchBizGroup(errC chan<- error) {
	r.flog.Info("RBACController start watch bizGroup")
	flog := r.flog.WithField("thread", "bizgroup")

	groupCoder := datasource.GetCoder(string(iam.BusinessGroupKind))
	if groupCoder == nil {
		errC <- fmt.Errorf("(%s) %s", iam.BusinessGroupKind, "coder not exist")
	}
	groupWatchChan := datasource.NewWatch(groupCoder)

	var version = "0"
	dbList, err := r.getDatabase()
	if err != nil {
		errC <- err
	}
	for _, s := range dbList {
		rawGroups, err := r.List(s, common.BUSINESSGROUP, "", false)
		if err != nil {
			errC <- err
		}

		for _, item := range rawGroups {
			group := &iam.BusinessGroup{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, group); err != nil {
				flog.Infof("reconcile bizGroup error %s\n", err)
			}

			if group.GetResourceVersion() > version {
				version = group.GetResourceVersion()
			}

			flog.Infof("get reconcile bizGroup %s\n", group.Name)
			if err := r.reconcileBizGroup(group); err != nil {
				flog.Infof("reconcile bizGroup error %s\n", err)
			}
		}
	}

	for _, s := range dbList {
		r.Watch(s, common.BUSINESSGROUP, version, groupWatchChan)
	}
	//r.Watch(common.LIZIDatabase, common.BUSINESSGROUP, version, groupWatchChan)
	for {
		select {
		case item, ok := <-groupWatchChan.ResultChan():
			if !ok {
				return
			}

			group := &iam.BusinessGroup{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, group); err != nil {
				flog.Infof("watch bizGroup unstruct error %s\n", err)
				continue
			}
			flog.Infof("watch reconcile bizGroup %s\n", group.Name)

			if err := r.reconcileBizGroup(group); err != nil {
				flog.Infof("watch bizGroup reconcile error %s\n", err)
			}
		}
	}
}

func (r *RBACController) WatchRole(errC chan<- error) {
	r.flog.Info("RBACController start watch role")
	flog := r.flog.WithField("thread", "role")

	roleCoder := datasource.GetCoder(string(rbac.RoleKind))
	if roleCoder == nil {
		errC <- fmt.Errorf("(%s) %s", rbac.RoleKind, "coder not exist")
	}
	roleWatchChan := datasource.NewWatch(roleCoder)

	var version = "0"
	dbList, err := r.getDatabase()
	if err != nil {
		errC <- err
	}
	for _, s := range dbList {
		rawRoles, err := r.List(s, common.ROLE, "", false)
		if err != nil {
			errC <- err
		}

		for _, item := range rawRoles {
			role := &rbac.Role{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, role); err != nil {
				flog.Infof("reconcile role error %s\n", err)
			}

			if role.GetResourceVersion() > version {
				version = role.GetResourceVersion()
			}

			flog.Infof("get reconcile role %s\n", role.Name)
			if err := r.reconcileRole(role); err != nil {
				flog.Infof("reconcile role error %s\n", err)
			}
		}
	}

	for _, s := range dbList {
		r.Watch(s, common.ROLE, version, roleWatchChan)
	}
	for {
		select {
		case item, ok := <-roleWatchChan.ResultChan():
			if !ok {
				return
			}

			role := &rbac.Role{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, role); err != nil {
				flog.Infof("watch role unstruct error %s\n", err)
				continue
			}
			flog.Infof("watch reconcile role %s\n", role.Name)

			if err := r.reconcileRole(role); err != nil {
				flog.Infof("watch role reconcile error %s\n", err)
			}
		}
	}
}

func (r *RBACController) WatchWorkspace(errC chan<- error) {
	r.flog.Info("RBACController start watch workspace")
	flog := r.flog.WithField("thread", "workspace")

	workspaceCoder := datasource.GetCoder(string(system.WorkspaceKind))
	if workspaceCoder == nil {
		errC <- fmt.Errorf("(%s) %s", system.WorkspaceKind, "coder not exist")
	}
	workspaceWatchChan := datasource.NewWatch(workspaceCoder)

	var version = "0"
	dbList, err := r.getDatabase()
	if err != nil {
		errC <- err
	}
	for _, s := range dbList {
		rawWorkspace, err := r.List(s, common.WORKSPACE, "", false)
		if err != nil {
			errC <- err
		}

		for _, item := range rawWorkspace {
			workspace := &system.Workspace{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, workspace); err != nil {
				flog.Infof("reconcile workspace error %s\n", err)
			}

			if workspace.GetResourceVersion() > version {
				version = workspace.GetResourceVersion()
			}

			flog.Infof("get reconcile workspace %s\n", workspace.Name)
			if err := r.reconcileWorkspace(workspace); err != nil {
				flog.Infof("reconcile workspace error %s\n", err)
			}
		}
	}

	for _, s := range dbList {
		r.Watch(s, common.WORKSPACE, version, workspaceWatchChan)
	}

	for {
		select {
		case item, ok := <-workspaceWatchChan.ResultChan():
			if !ok {
				return
			}

			workspace := &system.Workspace{}
			if err := obj.UnstructuredObjectToInstanceObj(&item, workspace); err != nil {
				flog.Infof("watch workspace %s unstructured error %s\n", workspace.GetName(), err)
				continue
			}
			flog.Infof("watch reconcile workspace %s\n", workspace.Name)

			if err := r.reconcileWorkspace(workspace); err != nil {
				flog.Infof("watch workspace %s reconcile error %s\n", workspace.GetName(), err)
			}
		}
	}
}

func (r *RBACController) LoopLiZiData(errC chan<- error) {
	r.flog.Infof("LoopLiZiData start")
	feishuObj := feishu.FeiShu{}
	for {
		if feishuObj.ExpireTime.Before(time.Now()) {
			if err := feishuObj.GetTenantAccessToken(); err != nil {
				errC <- err
			}
		}
		err := r.loopDepartment(feishuObj.TenantAccessToken, "0", "")
		if err != nil {
			errC <- err
		}
		time.Sleep(time.Minute * 60)
	}
}

func (r *RBACController) getDatabase() ([]string, error) {
	// 每个租户的数据都放在其自己的数据库下，通过租户获取所有数据库
	rawData, err := r.List(common.DefaultDatabase, common.TENANT, "", true)
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
