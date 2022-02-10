package system

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const SecurityGroupKind core.Kind = "securitygroup"

type IpProtocolType string

const (
	TCP  IpProtocolType = "TCP"
	UDP  IpProtocolType = "UDP"
	ICMP IpProtocolType = "ICMP"
	GRE  IpProtocolType = "GRE"
	ALL  IpProtocolType = "ALL"
)

type SecurityGroupRole struct {
	PortRange    string         `json:"port_range" bson:"port_range"`
	IpProtocol   IpProtocolType `json:"ip_protocol" bson:"ip_protocol"`
	SourceCidrIp string         `json:"source_cidr_ip" bson:"source_cidr_ip"`
}

type SecurityGroupSpec struct {
	LocalName string              `json:"local_name" bson:"local_name"`
	RegionId  string              `json:"region_id" bson:"region_id"`
	VpcId     string              `json:"vpc_id" bson:"vpc_id"`
	ID        string              `json:"id" bson:"id"`
	Status    string              `json:"status" bson:"status"`
	Ingress   []SecurityGroupRole `json:"ingress" bson:"ingress"`
	Egress    []SecurityGroupRole `json:"egress" bson:"egress"`
	Message   string              `json:"message" bson:"message"`
}

type SecurityGroup struct {
	core.Metadata `json:"metadata"`
	Spec          SecurityGroupSpec `json:"spec"`
}

func (v *SecurityGroup) Clone() core.IObject {
	result := &SecurityGroup{}
	core.Clone(v, result)
	return result
}

func (*SecurityGroup) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &SecurityGroup{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type SecurityGroupList struct {
	core.Metadata `json:"metadata"`
	Items         []SecurityGroup `json:"items"`
}

func (v *SecurityGroupList) GenerateListVersion() {
	var maxVersion string
	for _, item := range v.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	v.Metadata = core.Metadata{
		Kind:    "securityGroupList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(SecurityGroupKind), &SecurityGroup{})
}
