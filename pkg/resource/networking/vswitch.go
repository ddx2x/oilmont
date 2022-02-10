package networking

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const VSwitchKind core.Kind = "vswitch"

type VSwitchSpec struct {
	LocalName string `json:"local_name" bson:"local_name"`
	IP        string `json:"ip" bson:"ip"`
	Mask      string `json:"mask" bson:"mask"`
	Region    string `json:"region" bson:"region"`
	Zone      string `json:"zone" bson:"zone"`
	Id        string `json:"id" bson:"id"`
	VpcId     string `json:"vpc_id" bson:"vpc_id"`
	Status    string `json:"status" bson:"status"`
	Describe  string `json:"describe" bson:"describe"`
	Message   string `json:"message" bson:"message"`
}

type Vswitch struct {
	core.Metadata `json:"metadata"`
	Spec          VSwitchSpec `json:"spec"`
}

func (v *Vswitch) Clone() core.IObject {
	result := &Vswitch{}
	core.Clone(v, result)
	return result
}

func (*Vswitch) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Vswitch{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type VSwitchList struct {
	core.Metadata `json:"metadata"`
	Items         []Vswitch `json:"items"`
}

func (v *VSwitchList) GenerateListVersion() {
	var maxVersion string
	for _, item := range v.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	v.Metadata = core.Metadata{
		Kind:    "vSwitchList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(VSwitchKind), &Vswitch{})
}
