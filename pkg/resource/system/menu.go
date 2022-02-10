package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const MenuKind core.Kind = "menu"

type MenuType = string

const (
	Product MenuType = "product"
	Action  MenuType = "action"
)

type Menu struct {
	core.Metadata `json:"metadata"`
	Spec          MenuSpec `json:"spec"`
}

type MenuSpec struct {
	Link      string   `json:"link"`
	Title     string   `json:"title"`
	Icon      string   `json:"icon"`
	Level     uint8    `json:"level" bson:"level"`
	Type      MenuType `json:"type"`
	Parent    string   `json:"parent" bson:"parent"`
	IsSubMenu bool     `json:"is_sub_menu" bson:"is_sub_menu"`
}

func (o *Menu) Clone() core.IObject {
	result := &Menu{}
	core.Clone(o, result)
	return result
}

func (o *Menu) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Menu{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

func init() {
	datasource.RegistryCoder(string(MenuKind), &Menu{})
}
