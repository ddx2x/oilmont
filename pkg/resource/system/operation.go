package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const OperationKind core.Kind = "operation"

type OperationSpec struct {
	OP     string `json:"op"`
	Method string `json:"method" bson:"method"`
}

type Operation struct {
	core.Metadata `json:"metadata"`
	Spec          OperationSpec `json:"spec"`
}

func (op *Operation) Clone() core.IObject {
	result := &Operation{}
	core.Clone(op, result)
	return result
}

func (*Operation) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Operation{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type OperationList struct {
	core.Metadata `json:"metadata"`
	Items         []Operation `json:"items"`
}

func (o *OperationList) GenerateListVersion() {
	var maxVersion string
	for _, item := range o.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	o.Metadata = core.Metadata{
		Kind:    "operationList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(OperationKind), &Operation{})
}
