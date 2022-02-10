package userconfig

import (
	"github.com/ddx2x/oilmont/pkg/resource/iam"
)

type Cluster struct {
	Name string   `json:"name"`
	Ns   []string `json:"ns"`
}

type Az struct {
	Name      string     `json:"name"`
	LocalName string     `json:"localName"`
	Clusters  []*Cluster `json:"cluster"`
}

type Region struct {
	Name      string `json:"name"`
	LocalName string `json:"localName"`
	Azs       []*Az  `json:"azs"`
}

func (r *Region) GetAZ(localName string) *Az {
	for _, a := range r.Azs {
		if a.LocalName == localName {
			return a
		}
	}
	return nil
}

type Provider struct {
	Name       string    `json:"name"`
	LocalName  string    `json:"localName"`
	ThirdParty bool      `json:"thirdParty"`
	Regions    []*Region `json:"regions"`
}

func (p *Provider) GetRegion(localName string) *Region {
	for _, r := range p.Regions {
		if r.LocalName == localName {
			return r
		}
	}
	return nil
}

type UserMenu struct {
	Name     string      `json:"name"`
	Link     string      `json:"link"`
	Title    string      `json:"title"`
	Icon     string      `json:"icon"`
	Parent   bool        `json:"parent"`
	Children []*UserMenu `json:"children"`
}

type UserMenuTrees []*UserMenu

func (umt *UserMenuTrees) AddRoot(um *UserMenu) {
	*umt = append(*umt, um)
}

func (umt *UserMenuTrees) AddBranch(um *UserMenu, root string) {
	for _, u := range *umt {
		if u.Name != root {
			continue
		}
		u.Children = append(u.Children, um)
		break
	}
}

func (umt UserMenuTrees) AddLeaf(um *UserMenu, branch string) {
	for _, u := range umt {
		for _, children := range u.Children {
			if children.Name == branch {
				children.Children = append(children.Children, um)
			}
		}
	}
}

type Config struct {
	UserName          string                 `json:"userName"`
	Token             string                 `json:"token"`
	RoleType          iam.AccountType        `json:"roleType"`
	DefaultWorkspace  string                 `json:"defaultWorkspace"`
	AllowedWorkspaces []string               `json:"allowWorkspaces"`
	Providers         map[string]*Provider   `json:"providers"`
	ProductMenu       []string               `json:"productMenu"`
	ActionMenu        map[string][]string    `json:"actionMenu"`
	Menus             UserMenuTrees          `json:"menus"`
	Permission        map[string]interface{} `json:"permission"`
	Tenant            string                 `json:"tenant"`
	BizGroups         []string               `json:"bizGroups"`
	Roles             []string               `json:"roles"`
	IsTenantOwner     bool                   `json:"isTenantOwner"`
	OwnerBiz          []string               `json:"ownerBiz"`
}

func (cfg *Config) IsAdmin() bool {
	return cfg.RoleType == 3
}
