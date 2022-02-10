package cr

import (
	"fmt"

	"github.com/ddx2x/oilmont/pkg/core"
)

type CustomData struct {
	core.Metadata `json:"metadata"`
	Spec          map[string]interface{} `json:"spec"`
}

func (v *CustomData) Clone() core.IObject {
	result := &CustomData{}
	core.Clone(v, result)
	return result
}

func (*CustomData) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &CustomData{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type CustomDataList struct {
	core.Metadata `json:"metadata"`
	Items         []CustomData `json:"items"`
}

func (v *CustomDataList) GenerateListVersion() {
	var maxVersion string
	var resourceKind string

	for _, item := range v.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
		resourceKind = string(item.Kind)

	}
	v.Metadata = core.Metadata{
		Kind:    core.Kind(fmt.Sprintf("%sList", resourceKind)),
		Version: maxVersion,
	}
}

//
//func init() {
//	datasource.RegistryCoder(string(CustomDataKind), &CustomData{})
//}
