package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	AvailableZoneKind     core.Kind = "availablezone"
	AvailableZoneListKind core.Kind = "availablezoneList"
)

type AvailableZoneSpec struct {
	Region    string `json:"region" bson:"region"`
	LocalName string `json:"local_name" bson:"local_name"`
	ID        string `json:"id" bson:"id"`
}

type AvailableZone struct {
	core.Metadata `json:"metadata"`
	Spec          AvailableZoneSpec `json:"spec"`
}

func (i *AvailableZone) Clone() core.IObject {
	result := &AvailableZone{}
	core.Clone(i, result)
	return result
}

func (*AvailableZone) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &AvailableZone{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type AvailableZoneList struct {
	core.Metadata `json:"metadata"`
	Items         []AvailableZone `json:"items"`
}

func (a *AvailableZoneList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    AvailableZoneListKind,
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(AvailableZoneKind), &AvailableZone{})
}
