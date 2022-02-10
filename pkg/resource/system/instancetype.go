package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const InstanceTypeKind core.Kind = "instancetype"

type InstanceTypeSpec struct {
	Cores  int64  `json:"cores" bson:"cores"`
	Memory string `json:"memory" bson:"memory"`
	Region string `json:"region" bson:"region"`
	ID     string `json:"id" bson:"id"`
	Zone   string `json:"zone" bson:"zone"`
}

type InstanceType struct {
	core.Metadata `json:"metadata"`
	Spec          InstanceTypeSpec `json:"spec"`
}

func (r *InstanceType) Clone() core.IObject {
	result := &InstanceType{}
	core.Clone(r, result)
	return result
}

func (*InstanceType) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &InstanceType{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type InstanceTypeList struct {
	core.Metadata `json:"metadata"`
	Items         []InstanceType `json:"items"`
}

func (r *InstanceTypeList) GenerateListVersion() {
	var maxVersion string
	for _, item := range r.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}

	r.Metadata = core.Metadata{
		Kind:    "instanceTypeList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(InstanceTypeKind), &InstanceType{})
}
