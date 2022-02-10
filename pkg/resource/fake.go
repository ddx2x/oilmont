package resource

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const fakeKind = "fake"

var _ core.IObject = &Fake{}

type Fake struct {
	core.Metadata `json:"metadata" bson:"metadata"`

	Spec fakeSpec `json:"spec" bson:"spec"`
}

type fakeSpec struct{}

func (f *Fake) Clone() core.IObject {
	result := &Fake{}
	core.Clone(f, result)
	return result
}

func (*Fake) Decode(opData map[string]interface{}) (core.IObject, error) {
	f := &Fake{}
	if err := core.UnmarshalToIObject(opData, f); err != nil {
		return nil, err
	}
	return f, nil
}

func init() {
	datasource.RegistryCoder(string(fakeKind), &Fake{})
}
