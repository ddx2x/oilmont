package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	PorviderKind core.Kind = "provider"
)

type Provider struct {
	core.Metadata `json:"metadata"`
	Spec          ProviderSpec `json:"spec"`
}

type ProviderSpec struct {
	LocalName    string `json:"localName" bson:"local_name"`
	ThirdParty   bool   `json:"thirdParty"`
	AccessKey    string `json:"accessKey"`
	AccessSecret string `json:"accessSecret"`
}

func (i *Provider) Clone() core.IObject {
	result := &Provider{}
	core.Clone(i, result)
	return result
}

func (i *Provider) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Provider{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

func init() {
	datasource.RegistryCoder(string(PorviderKind), &Provider{})
}
