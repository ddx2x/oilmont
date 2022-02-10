package rbac

/*
	管理账户和权限的中间表
*/

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const AccountPermissionKind core.Kind = "accountpermission"

type AccountPermissionMap map[string]map[string]map[string]struct{} // {"workspace":{"resource-1":{"op-1":{},"op-2":{}}}}

type AccountPermissionSpec struct {
	Account    string               `json:"account"`
	Permission AccountPermissionMap `json:"permission"` //{"resource":["op1","op2"]}
	Menus      []string             `json:"menus" bson:"menus"`
	Workspaces []string             `json:"workspaces" bson:"workspaces"`
	Roles      []string             `json:"roles" bson:"roles"`
}

type AccountPermission struct {
	core.Metadata `json:"metadata"`
	Spec          AccountPermissionSpec `json:"spec"`
}

func (ap *AccountPermission) Clone() core.IObject {
	result := &AccountPermission{}
	core.Clone(ap, result)
	return result
}

func (ap *AccountPermission) Decode(m map[string]interface{}) (core.IObject, error) {
	panic("implement me")
}

func init() {
	datasource.RegistryCoder(string(AccountPermissionKind), &AccountPermission{})
}
