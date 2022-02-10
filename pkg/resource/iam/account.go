package iam

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/gtm"
)

const (
	AccountKind     core.Kind = "account"
	AccountListKind core.Kind = "accountList"
)

type AccountType uint8

const (
	AccountTypeAdmin AccountType = iota + 3
	AccountTypeTenant
)

type AccountSpec struct {
	BusinessGroupRole map[string][]string `json:"business_group_role" bson:"business_group_role"`

	AccountType `json:"account_type" bson:"account_type"`
	Password    string `json:"password"`
}

type Account struct {
	core.Metadata `json:"metadata"`
	Spec          AccountSpec `json:"spec"`
}

func (u *Account) Clone() core.IObject {
	result := &Account{}
	core.Clone(u, result)
	return result
}

func (*Account) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Account{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type AccountList struct {
	core.Metadata `json:"metadata"`
	Items         []Account `json:"items"`
}

func (*AccountList) Decode(op *gtm.Op) (core.IObject, error) {
	action := &AccountList{}
	if err := core.UnmarshalToIObject(op.Data, action); err != nil {
		return nil, err
	}
	return action, nil
}

func (a *AccountList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    "accountList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(AccountKind), &Account{})
}
