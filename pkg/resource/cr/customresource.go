package cr

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const CustomResourceKind core.Kind = "customresource"

type CustomResourceSpec struct {
	CustomResource map[string]string `json:"custom_resource" bson:"customResource"`
}

type CustomResource struct {
	core.Metadata `json:"metadata"`
	Spec          CustomResourceSpec `json:"spec"`
}

func (v *CustomResource) Clone() core.IObject {
	result := &CustomResource{}
	core.Clone(v, result)
	return result
}

func (*CustomResource) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &CustomResource{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type CustomResourceList struct {
	core.Metadata `json:"metadata"`
	Items         []CustomResource `json:"items"`
}

func (v *CustomResourceList) GenerateListVersion() {
	var maxVersion string
	for _, item := range v.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	v.Metadata = core.Metadata{
		Kind:    "customResourceList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(CustomResourceKind), &CustomResource{})
}
