package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	RegionKind     core.Kind = "region"
	RegionListKind core.Kind = "regionList"
)

type RegionSpec struct {
	LocalName string `json:"local_name" bson:"local_name"`
	Endpoint  string `json:"endpoint" bson:"endpoint"`
	ID        string `json:"id" bson:"id"`
}

type Region struct {
	core.Metadata `json:"metadata"`
	Spec          RegionSpec `json:"spec"`
}

func (i *Region) Clone() core.IObject {
	result := &Region{}
	core.Clone(i, result)
	return result
}

func (i *Region) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Region{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type RegionList struct {
	core.Metadata `json:"metadata"`
	Items         []Region `json:"items"`
}

func (a *RegionList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    RegionListKind,
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(RegionKind), &Region{})
}
