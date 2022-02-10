package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const ResourceKind core.Kind = "resource"

type Info struct {
	Price float32 `json:"price"`
	Data  string  `json:"data"`
}

type ResourceType int

const (
	ResourceTypePrivate ResourceType = iota + 1
	ResourceTypePublic
)

type ResourceTypeOfRole int

const (
	ResourceTypeOfRoleUser ResourceTypeOfRole = iota + 1
	ResourceTypeOfRoleTenantOwner
	ResourceTypeOfRoleBizOwner
)

type ResourceSpec struct {
	ResourceName string             `json:"resourceName" bson:"resourceName"` // eg virtualmachines
	Menu         string             `json:"menu" bson:"menu"`                 // eg account/帐户
	ApiVersion   string             `json:"apiVersion" bson:"apiVersion"`     // eg: virtualMachine --> vm.ddx2x.nip
	Group        string             `json:"group"`                            // eg: v1
	Kind         string             `json:"kind"`                             // VirtualMachine
	Ops          []string           `json:"ops" bson:"ops"`
	Info         Info               `json:"info"`
	Type         ResourceType       `json:"type" bson:"type"`
	TypeOfRole   ResourceTypeOfRole `json:"type_of_role" bson:"type_of_role"`
}

type Resource struct {
	core.Metadata `json:"metadata"`
	Spec          ResourceSpec `json:"spec"`
}

func (p *Resource) Clone() core.IObject {
	result := &Resource{}
	core.Clone(p, result)
	return result
}

func (*Resource) Decode(opData map[string]interface{}) (core.IObject, error) {
	object := &Resource{}
	if err := core.UnmarshalToIObject(opData, object); err != nil {
		return nil, err
	}
	return object, nil
}

func init() {
	datasource.RegistryCoder(string(ResourceKind), &Resource{})
}
