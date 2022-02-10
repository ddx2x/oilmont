package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	ThemeKind     core.Kind = "theme"
	ThemeListKind core.Kind = "themeList"
)

type ThemeSpec struct {
	MUI      string `json:"mui" bson:"mui"`
	DataGrid string `json:"data_grid" bson:"data_grid"`
	Color    string `json:"color" bson:"color"`
	Extend   string `json:"extend" bson:"extend"`
}

type Theme struct {
	core.Metadata `json:"metadata"`
	Spec          ThemeSpec `json:"spec"`
}

func (i *Theme) Clone() core.IObject {
	result := &Theme{}
	core.Clone(i, result)
	return result
}

func (i *Theme) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Theme{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type ThemeList struct {
	core.Metadata `json:"metadata"`
	Items         []Theme `json:"items"`
}

func (a *ThemeList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    ThemeListKind,
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(ThemeKind), &Theme{})
}
