package compute

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

const StorageKind core.Kind = "storage"

const (
	StorageStateCreating  = "creating"
	StorageStateAvailable = "available"
	StorageStateInUse     = "in-use"

	//aws
	StorageStateDeleting = "deleting"
	StorageStateDeleted  = "deleted"
	StorageStateError    = "error"

	//aws attach state
	StorageStateAttached = "attached"
	StorageStateDetached = "detached"
	StorageStateBusy     = "busy"

	//aliyun
	StorageStateAttaching = "attaching"
	StorageStateDetaching = "detaching"
	StorageStateReIniting = "reiniting"

	//需要 ctrl 处理的 state
	StorageStateCreate = "create"
	StorageStateAttach = "attach"
	StorageStateDetach = "detach"
	StorageStateDelete = "delete"
)

type Attachment struct {
	AttachedTime string `json:"attached_time" bson:"attached_time"`
	Device       string `json:"device" bson:"device"`
	InstanceId   string `json:"instance_id" bson:"instance_id"`
	State        string `json:"state" bson:"state"`
}

type StorageSpec struct {
	Attachments        []Attachment `json:"attachments" bson:"attachments"`
	CategoryType       string       `json:"category_type" bson:"category_type"`
	DeleteWithInstance bool         `json:"delete_with_instance" bson:"delete_with_instance"`
	Description        string       `json:"description" bson:"description"`
	DiskChargeType     string       `json:"disk_charge_type" bson:"disk_charge_type"`
	DiskType           string       `json:"disk_type" bson:"disk_type"`
	IOPS               int          `json:"iops" bson:"iops"`
	Throughput         int          `json:"throughput" bson:"throughput"`
	LocalName          string       `json:"local_name" bson:"local_name"`
	Message            string       `json:"message" bson:"message"`
	Region             string       `json:"region" bson:"region"`
	Zone               string       `json:"zone" bson:"zone"`
	Size               int          `json:"size" bson:"size"`
	State              string       `json:"state" bson:"state"`
	Status             string       `json:"status" bson:"status"`
	StorageId          string       `json:"storage_id" bson:"storage_id"`
}

type Storage struct {
	core.Metadata `json:"metadata"`
	Spec          StorageSpec `json:"spec"`
}

func (s *Storage) Clone() core.IObject {
	result := &Storage{}
	core.Clone(s, result)
	return result
}

func (*Storage) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &Storage{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type StorageList struct {
	core.Metadata `json:"metadata"`
	Items         []Storage `json:"items"`
}

func (a *StorageList) GenerateListVersion() {
	var maxVersion string
	for _, item := range a.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	a.Metadata = core.Metadata{
		Kind:    "storageList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(StorageKind), &Storage{})
}
