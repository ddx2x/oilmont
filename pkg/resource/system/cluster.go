package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const (
	ClusterKind     core.Kind = "cluster"
	ClusterListKind core.Kind = "clusterList"
)

type ClusterSpec struct {
	Config     map[string]interface{} `json:"config"`
	Provider   string                 `json:"provider"`
	Region     string                 `json:"region"`
	Az         string                 `json:"az"`
	Namespaces []string               `json:"namespaces" bson:"namespaces"`
	Nodes      []string               `json:"nodes" bson:"nodes"`
}

type Cluster struct {
	core.Metadata `json:"metadata"`
	Spec          ClusterSpec `json:"spec"`
}

func (i *Cluster) Clone() core.IObject {
	result := &Cluster{}
	core.Clone(i, result)
	return result
}

func (*Cluster) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Cluster{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type ClusterList struct {
	core.Metadata `json:"metadata"`
	Items         []Cluster `json:"items"`
}

func (a *ClusterList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    ClusterListKind,
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(ClusterKind), &Cluster{})
}
