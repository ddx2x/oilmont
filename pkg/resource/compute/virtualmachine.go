package compute

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const VirtualMachineKind core.Kind = "virtualmachine"

type VirtualMachineType string

type VirtualMachineStateType string

const (
	Running      VirtualMachineStateType = "running"
	Starting     VirtualMachineStateType = "starting"
	Stopped      VirtualMachineStateType = "stopped"
	Stopping     VirtualMachineStateType = "stopping"
	Restarting   VirtualMachineStateType = "restarting"
	Deleting     VirtualMachineStateType = "deleting"
	Updating     VirtualMachineStateType = "updating"
	Ready        VirtualMachineStateType = "ready"
	Unknown      VirtualMachineStateType = "unknown"
	Provisioning VirtualMachineStateType = "Provisioning"
)

type VmStorage struct {
	Name   string `json:"name" bson:"name"`
	Status string `json:"status" bson:"status"`

	// kubeVirt
	Quantity string `json:"quantity" bson:"quantity"`
	Registry string `json:"registry" bson:"registry"`
	Type     string `json:"type" bson:"type"`

	// thirdProvider
	DeleteOnTermination bool   `json:"delete_on_termination" bson:"delete_on_termination"`
	Iops                int64  `json:"iops" bson:"iops"`
	Throughput          int64  `json:"throughput" bson:"throughput"`
	VolumeId            string `json:"volume_id" bson:"volume_id"`
	VolumeSize          int64  `json:"volume_size" bson:"volume_size"`
	VolumeType          string `json:"volume_type" bson:"volume_type"`
	DiskType            string `json:"disk_type" bson:"disk_type"`
	State               string `json:"state" bson:"state"`
	Message             string `json:"message" bson:"message"`
}

type NetWorkInterface struct {
	NetworkInterfaceId string   `json:"network_interface_id" bson:"network_interface_id"`
	PrivateIp          string   `json:"private_ip" bson:"private_ip"`
	SecondaryPrivateIp []string `json:"secondary_private_ip" bson:"secondary_private_ip"`
	Mac                string   `json:"mac" bson:"mac"`
	PublicIp           string   `json:"public_ip" bson:"public_ip"`
	SecureGroups       []string `json:"secure_groups" bson:"secure_groups"`
	VpcId              string   `json:"vpc_id" bson:"vpc_id"`
	VSwitchId          string   `json:"vswitch_id" bson:"vswitch_id"`
	Type               string   `json:"type" bson:"type"`
	State              string   `json:"state" bson:"state"`
	Status             string   `json:"status" bson:"status"`
	Message            string   `json:"message" bson:"message"`
}

type VirtualMachineSpec struct {
	CPU              string                  `json:"cpu" bson:"cpu"`
	Memory           string                  `json:"memory" bson:"memory"`
	PublicIpAddress  []string                `json:"public_ip_address" bson:"public_ip_address"`
	PrivateIpAddress []string                `json:"private_ip_address" bson:"private_ip_address"`
	LocalName        string                  `json:"local_name" bson:"local_name"`
	RegionId         string                  `json:"region_id" bson:"region_id"`
	Az               string                  `json:"az"`
	InstanceId       string                  `json:"instance_id" bson:"instance_id"`
	InstanceType     string                  `json:"instance_type" bson:"instance_type"`
	VPCId            string                  `json:"vpc_id" bson:"vpc_id"`
	VSwitchId        string                  `json:"vswitch_id" bson:"vswitch_id"`
	ImageId          string                  `json:"image_id" bson:"image_id"`
	SecurityGroup    []string                `json:"security_group" bson:"security_group"`
	Vendor           string                  `json:"vendor" bson:"vendor"`
	RootDevice       string                  `json:"root_device" bson:"root_device"`
	Status           string                  `json:"status" bson:"status"` // ddx2x与k8s同步状态
	State            VirtualMachineStateType `json:"state" bson:"state"`   // state是vm状态
	Message          string                  `json:"message" bson:"message"`
	Storage          []VmStorage             `json:"storage" bson:"storage"`
	NetWorkInterface []NetWorkInterface      `json:"network_interface" bson:"network_interface"`
	CreateTime       string                  `json:"create_time" bson:"create_time"`
	Os               string                  `json:"os" bson:"os"`
}

type VirtualMachine struct {
	core.Metadata `json:"metadata"`
	Spec          VirtualMachineSpec `json:"spec"`
}

func (v *VirtualMachine) Clone() core.IObject {
	result := &VirtualMachine{}
	core.Clone(v, result)
	return result
}

func (*VirtualMachine) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &VirtualMachine{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type VirtualMachineList struct {
	core.Metadata `json:"metadata"`
	Items         []VirtualMachine `json:"items"`
}

func (a *VirtualMachineList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    "virtualMachineList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(VirtualMachineKind), &VirtualMachine{})
}
