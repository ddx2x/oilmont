package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	WorkspaceKind     core.Kind = "workspace"
	WorkspaceListKind core.Kind = "workspaceList"
)

type WorkspaceSpec struct {
	Tenant string `json:"tenant" bson:"tenant"`
}

type Workspace struct {
	core.Metadata `json:"metadata"`
	Spec          WorkspaceSpec `json:"spec"`
}

func (i *Workspace) Clone() core.IObject {
	result := &Workspace{}
	core.Clone(i, result)
	return result
}

func (i *Workspace) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Workspace{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type WorkspaceList struct {
	core.Metadata `json:"metadata"`
	Items         []Workspace `json:"items"`
}

func (a *WorkspaceList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    WorkspaceListKind,
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(WorkspaceKind), &Workspace{})
}
