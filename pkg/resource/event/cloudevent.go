package event

import (
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
)

type OperatorStatusType = string

const (
	CloudEventSuccess = "success"
	CloudEventReject  = "reject"
	CloudEventFail    = "fail"
)

const (
	CloudEventKind     core.Kind = "cloudevent"
	CloudEventListKind core.Kind = "cloudeventList"
)

type CloudEventSpec struct {
	Source          string             `json:"source" bson:"source"`
	Name            string             `json:"name" bson:"name"`
	Action          string             `json:"action" bson:"action"`
	Operator        string             `json:"operator" bson:"operator"`
	SourceWorkspace string             `json:"source_workspace" bson:"source_workspace"`
	OperatorStatus  OperatorStatusType `json:"operator_status" bson:"operator_status"`
}

type CloudEvent struct {
	core.Metadata `json:"metadata" bson:"metadata"`
	Spec          CloudEventSpec `json:"spec" bson:"spec"`
}

func (c *CloudEvent) Clone() core.IObject {
	result := &CloudEvent{}
	core.Clone(c, result)
	return result
}

func (*CloudEvent) Decode(opData map[string]interface{}) (core.IObject, error) {
	action := &CloudEvent{}
	if err := core.UnmarshalToIObject(opData, action); err != nil {
		return nil, err
	}
	return action, nil
}

type CloudEventList struct {
	core.Metadata `json:"metadata"`
	Items         []CloudEvent `json:"items"`
}

func (v *CloudEventList) GenerateListVersion() {
	var maxVersion string
	for _, item := range v.Items {
		if item.Version > maxVersion {
			maxVersion = item.Version
		}
	}
	v.Metadata = core.Metadata{
		Kind:    "cloudeventList",
		Version: maxVersion,
	}
}

func init() {
	datasource.RegistryCoder(string(CloudEventKind), &CloudEvent{})
}
