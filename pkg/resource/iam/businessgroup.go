package iam

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const BusinessGroupKind core.Kind = "businessgroup"

type BusinessGroupSpec struct {
	Owner     string   `json:"owner" bson:"owner"`
	OwnerName string   `json:"ownerName"`
	Roles     []string `json:"roles" bson:"roles"`
}

type BusinessGroup struct {
	core.Metadata `json:"metadata"`
	Spec          BusinessGroupSpec `json:"spec"`
}

func (g *BusinessGroup) Clone() core.IObject {
	result := &BusinessGroup{}
	core.Clone(g, result)
	return result
}

func (*BusinessGroup) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &BusinessGroup{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type BusinessGroupList struct {
	core.Metadata `json:"metadata"`
	Items         []BusinessGroup `json:"items"`
}

func (b *BusinessGroupList) GenerateListVersion() {
	var maxVersion string
	for _, item := range b.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	b.Metadata = core.Metadata{
		Kind:    "businessGroupList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(BusinessGroupKind), &BusinessGroup{})
}
