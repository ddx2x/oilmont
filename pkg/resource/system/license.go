package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

type SshType string

const (
	LicenseKind     core.Kind = "license"
	LicenseListKind core.Kind = "licenseList"

	Password SshType = "password"
	RSA      SshType = "rsa"
)

type LicenseSpec struct {
	Vendor        string  `json:"vendor" bson:"vendor"`
	Region        string  `json:"region" bson:"region"`
	AvailableZone string  `json:"available_zone" bson:"available_zone"`
	SshType       SshType `json:"ssh_type" bson:"ssh_type"`
	Key           string  `json:"key" bson:"key"`
}

type License struct {
	core.Metadata `json:"metadata"`
	Spec          LicenseSpec `json:"spec"`
}

func (i *License) Clone() core.IObject {
	result := &License{}
	core.Clone(i, result)
	return result
}

func (*License) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &License{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type LicenseList struct {
	core.Metadata `json:"metadata"`
	Items         []License `json:"items"`
}

func (l *LicenseList) GenerateListVersion() {
	var maxVersion string
	for _, item := range l.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	l.Metadata = core.Metadata{
		Kind:    LicenseListKind,
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(LicenseKind), &License{})
}
