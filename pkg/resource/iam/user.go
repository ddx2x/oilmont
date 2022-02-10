package iam

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const UserKind core.Kind = "user"

type Avatar struct {
	Avatar240    string `json:"avatar_240" bson:"avatar_240"`
	Avatar640    string `json:"avatar_640" bson:"avatar_640"`
	Avatar72     string `json:"avatar_72" bson:"avatar_72"`
	AvatarOrigin string `json:"avatar_origin" bson:"avatar_origin"`
}

type UserSpec struct {
	EnName  string `json:"en_name" bson:"en_name"`
	CnName  string `json:"cn_name" bson:"cn_name"`
	Email   string `json:"email" bson:"email"`
	Account string `json:"account" bson:"account"`
	Avatar  Avatar `json:"avatar" bson:"avatar"`
	OpenID  string `json:"open_id" bson:"open_id"`
}

type User struct {
	core.Metadata `json:"metadata"`
	Spec          UserSpec `json:"spec"`
}

func (r *User) Clone() core.IObject {
	result := &User{}
	core.Clone(r, result)
	return result
}

func (*User) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &User{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type UserList struct {
	core.Metadata `json:"metadata"`
	Items         []User `json:"items"`
}

func (r *UserList) GenerateListVersion() {
	var maxVersion string
	for _, item := range r.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}

	r.Metadata = core.Metadata{
		Kind:    "userList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(UserKind), &User{})
}
