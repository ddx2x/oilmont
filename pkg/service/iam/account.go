package iam

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/service"
	"golang.org/x/crypto/bcrypt"
)

type AccountService struct {
	service.IService
}

func NewAccount(i service.IService) *AccountService {
	return &AccountService{i}
}

func (as *AccountService) List(tenant, workspace string) (*iam.AccountList, error) {
	filter := map[string]interface{}{}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}
	data := make([]iam.Account, 0)
	err := as.IService.ListToObject(tenant, common.ACCOUNT, filter, &data, true)
	if err != nil {
		return nil, err
	}

	accountList := &iam.AccountList{Items: data}
	accountList.GenerateListVersion()

	return accountList, nil
}

func (as *AccountService) GetByName(tenant, workspace, name string) (*iam.Account, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if workspace != "" {
		filter[common.FilterWorkspace] = workspace
	}

	account := &iam.Account{}
	err := as.IService.GetByFilter(tenant, common.ACCOUNT, account, filter, true)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (as *AccountService) Create(reqAccount *iam.Account) (core.IObject, error) {
	if _, err := as.GetByName(reqAccount.GetTenant(), reqAccount.GetWorkspace(), reqAccount.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("account exists")
	}

	if reqAccount.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if reqAccount.Spec.Password == "" {
		reqAccount.Spec.Password = uuid.New().String()
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(reqAccount.Spec.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	reqAccount.Spec.Password = string(hash)
	reqAccount.Kind = iam.AccountKind
	reqAccount.GenerateVersion()

	_, err = as.IService.Create(reqAccount.GetTenant(), common.ACCOUNT, reqAccount)
	if err != nil {
		return nil, err
	}
	return reqAccount, nil
}

func (as *AccountService) Update(tenant, workspace, name string, reqAccount *iam.Account, paths ...string) (core.IObject, bool, error) {
	obj, update, err := as.IService.Apply(
		tenant,
		common.ACCOUNT,
		name,
		reqAccount,
		false,
		paths...)
	if err != nil {
		return nil, false, err
	}

	return obj, update, nil
}

func (as *AccountService) Delete(tenant, workspace, name string) (core.IObject, error) {
	account, err := as.GetByName(tenant, workspace, name)
	if err != nil {
		return nil, err
	}
	err = as.DeleteObject(tenant, common.ACCOUNT, account.Name, account, true)
	return account, err
}

func (as *AccountService) CheckCreateAccountWorkspacePermission(userName string, workspace string) bool {
	account := iam.Account{}
	accountFilter := map[string]interface{}{"metadata.name": userName}

	if err := as.GetByFilter(common.DefaultDatabase, common.ACCOUNT, &account, accountFilter, true); err != nil {
		return false
	}

	switch account.Spec.AccountType {
	case iam.AccountTypeAdmin:
		return true

	}

	return false
}

func (as *AccountService) CheckCreateAccountTypePermission(userName string, accountType iam.AccountType) bool {
	account := iam.Account{}
	accountFilter := map[string]interface{}{"metadata.name": userName}

	if err := as.GetByFilter(common.DefaultDatabase, common.ACCOUNT, &account, accountFilter, true); err != nil {
		return false
	}

	if account.Spec.AccountType == iam.AccountTypeAdmin {
		return true
	}

	return true
}
