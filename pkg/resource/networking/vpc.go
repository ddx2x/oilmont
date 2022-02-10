package networking

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const VirtualPrivateCloudKind core.Kind = "vpc"

type VirtualPrivateCloudType string

const (
	Vpc VirtualPrivateCloudType = "Vpc"
)

type VirtualPrivateCloudTypeStateType string

type VirtualPrivateCloudSpec struct {
	LocalName string `json:"local_name" bson:"local_name"`
	IP        string `json:"ip" bson:"ip"`
	Mask      string `json:"mask" bson:"mask"`
	Region    string `json:"region" bson:"region"`
	Status    string `json:"status" bson:"status"`
	ID        string `json:"id" bson:"id"`
	Message   string `json:"message" bson:"message"`
}

type VirtualPrivateCloud struct {
	core.Metadata `json:"metadata"`
	Spec          VirtualPrivateCloudSpec `json:"spec"`
}

func (v *VirtualPrivateCloud) Clone() core.IObject {
	result := &VirtualPrivateCloud{}
	core.Clone(v, result)
	return result
}

func (*VirtualPrivateCloud) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &VirtualPrivateCloud{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type VirtualPrivateCloudList struct {
	core.Metadata `json:"metadata"`
	Items         []VirtualPrivateCloud `json:"items"`
}

func (v *VirtualPrivateCloudList) GenerateListVersion() {
	var maxVersion string
	for _, item := range v.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	v.Metadata = core.Metadata{
		Kind:    "vpcList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(VirtualPrivateCloudKind), &VirtualPrivateCloud{})
}
