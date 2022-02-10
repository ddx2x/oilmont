package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	RelationKind core.Kind = "relation"
)

type RelationSpec struct {
	RelationKind string            `json:"relation_kind" bson:"relation_kind"`
	Resources    map[string]string `json:"resources" bson:"resources"`
}

type Relation struct {
	core.Metadata `json:"metadata"`
	Spec          RelationSpec `json:"spec"`
}

func (u *Relation) Clone() core.IObject {
	result := &Relation{}
	core.Clone(u, result)
	return result
}

func (*Relation) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Relation{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

func init() {
	datasource.RegistryCoder(string(RelationKind), &Relation{})
}
