package networking

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const NetworkInterfaceKind core.Kind = "networkinterface"

type ENIAttachment struct {
	AttachTime string `json:"attach_time" bson:"attach_time"`
	InstanceId string `json:"instance_id" bson:"instance_id"`
	Status     string `json:"status" bson:"status"`
}

type ENIPrivateIpSet struct {
	Primary          bool   `json:"primary" bson:"primary"`
	PrivateDnsName   string `json:"private_dns_name" bson:"private_dns_name"`
	PrivateIpAddress string `json:"private_ip_address" bson:"private_ip_address"`
	PublicIpAddress  string `json:"public_ip_address" bson:"public_ip_address"`
}

type NetworkInterfaceSpec struct {
	Attachment       ENIAttachment     `json:"attachment" bson:"attachment"`
	Description      string            `json:"description" bson:"description"`
	MacAddress       string            `json:"mac_address" bson:"mac_address"`
	PrivateDnsName   string            `json:"private_dns_name" bson:"private_dns_name"`
	PrivateIpAddress string            `json:"private_ip_address" bson:"private_ip_address"`
	PrivateIpSets    []ENIPrivateIpSet `json:"private_ip_sets" bson:"private_ip_sets"`
	PublicDnsName    string            `json:"public_dns_name" bson:"public_dns_name"`
	PublicIpAddress  string            `json:"public_ip_address" bson:"public_ip_address"`
	SecurityGroupIds []string          `json:"security_group_ids" bson:"security_group_ids"`
	Ipv6             []string          `json:"ipv6" bson:"ipv6"`
	SubnetId         string            `json:"subnet_id" bson:"subnet_id"`
	VPCId            string            `json:"vpc_id" bson:"vpc_id"`
	Type             string            `json:"type" bson:"type"`
	ID               string            `json:"id" bson:"id"`
	LocalName        string            `json:"local_name" bson:"local_name"`
	Region           string            `json:"region" bson:"region"`
	Zone             string            `json:"zone" bson:"zone"`
	Status           string            `json:"status" bson:"status"`
	State            string            `json:"state" bson:"state"`
	Message          string            `json:"message" bson:"message"`
}

type NetworkInterface struct {
	core.Metadata `json:"metadata"`
	Spec          NetworkInterfaceSpec `json:"spec"`
}

func (v *NetworkInterface) Clone() core.IObject {
	result := &NetworkInterface{}
	core.Clone(v, result)
	return result
}

func (*NetworkInterface) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &NetworkInterface{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type NetworkInterfaceList struct {
	core.Metadata `json:"metadata"`
	Items         []NetworkInterface `json:"items"`
}

func (v *NetworkInterfaceList) GenerateListVersion() {
	var maxVersion string
	for _, item := range v.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	v.Metadata = core.Metadata{
		Kind:    "networkInterfaceList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(NetworkInterfaceKind), &NetworkInterface{})
}
