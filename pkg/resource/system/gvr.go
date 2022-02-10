package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	GVRKind     core.Kind = "gvr"
	GVRListKind core.Kind = "gvrList"
)

type GVRSpec struct {
	Kind    string `json:"kind"`
	Group   string `json:"group"`
	Version string `json:"version"`
}

type GVR struct {
	core.Metadata `json:"metadata"`
	Spec          GVRSpec `json:"spec"`
}

func (r *GVR) Clone() core.IObject {
	result := &GVR{}
	core.Clone(r, result)
	return result
}

func (*GVR) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &GVR{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

func init() {
	datasource.RegistryCoder(string(GVRKind), &GVR{})
}
