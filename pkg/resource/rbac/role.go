package rbac

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const RoleKind core.Kind = "role"

type RoleSpec struct {
	Business   string                 `json:"business"`
	Remark     string                 `json:"remark"`
	Permission map[string]interface{} `json:"permission"`
}

type Role struct {
	core.Metadata `json:"metadata"`
	Spec          RoleSpec `json:"spec"`
}

func (r *Role) Clone() core.IObject {
	result := &Role{}
	core.Clone(r, result)
	return result
}

func (*Role) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Role{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type RoleList struct {
	core.Metadata `json:"metadata"`
	Items         []Role `json:"items"`
}

func (r *RoleList) GenerateListVersion() {
	var maxVersion string
	for _, item := range r.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}

	r.Metadata = core.Metadata{
		Kind:    "roleList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(RoleKind), &Role{})
}
