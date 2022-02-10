package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const TenantKind core.Kind = "tenant"

type ReqTenantSpec struct {
	Owner string `json:"owner" bson:"owner"`
	Type  string `json:"type" bson:"type"`
	Key   string `json:"key" bson:"key"`
}

type ReqTenant struct {
	core.Metadata `json:"metadata"`
	Spec          ReqTenantSpec `json:"spec"`
}

type TenantSpec struct {
	Owner      string                 `json:"owner" bson:"owner"`
	Type       string                 `json:"type" bson:"type"`
	Key        string                 `json:"key" bson:"key"`
	Allowed    map[string]interface{} `json:"allowed"`
	Permission map[string]interface{} `json:"permission"`
	Menus      map[string]interface{} `json:"menus"` // {"SDN":{"TOP":["underlay"]}}
	Clusters   []string               `json:"clusters" bson:"clusters"`
}

type Tenant struct {
	core.Metadata `json:"metadata"`
	Spec          TenantSpec `json:"spec"`
}

func (t *Tenant) Clone() core.IObject {
	result := &Tenant{}
	core.Clone(t, result)
	return result
}

func (*Tenant) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Tenant{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type TenantList struct {
	core.Metadata `json:"metadata"`
	Items         []Tenant `json:"items"`
}

func (r *TenantList) GenerateListVersion() {
	var maxVersion string
	for _, item := range r.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}

	r.Metadata = core.Metadata{
		Kind:    "tenantList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(TenantKind), &Tenant{})
}
