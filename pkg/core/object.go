package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddx2x/oilmont/pkg/utils/uuid"
)

type Kind string

// AreaType cloud platform deploy geo flag
type AreaType = uint32

type Metadata struct {
	Name      string                 `json:"name" bson:"name"`
	Kind      Kind                   `json:"kind"  bson:"kind"`
	Version   string                 `json:"version" bson:"version"`
	UUID      string                 `json:"uuid" bson:"uuid"`
	IsDelete  bool                   `json:"is_delete" bson:"is_delete"`
	Tenant    string                 `json:"tenant" bson:"tenant"`
	Namespace string                 `json:"namespace" bson:"namespace"`
	Workspace string                 `json:"workspace" bson:"workspace"`
	Labels    map[string]interface{} `json:"labels" bson:"labels"`
	Area      AreaType               `json:"area" bson:"area"`
}

func (m *Metadata) GetMateData() Metadata {
	return *m
}

func (m *Metadata) Delete() { m.IsDelete = true }

func (m *Metadata) Clone() IObject { panic("implement me") }

func (m *Metadata) SetLabel(key string, value interface{}) {
	if m.Labels == nil {
		m.Labels = make(map[string]interface{})
	}
	m.Labels[key] = value
}

func (m *Metadata) GetUUID() string {
	return m.UUID
}

func (m *Metadata) GetResourceVersion() string {
	return m.Version
}

func (m *Metadata) GetName() string {
	return m.Name
}

func (m *Metadata) GetKind() Kind {
	return m.Kind
}

func (m *Metadata) SetKind(k Kind) {
	m.Kind = k
}

func (m *Metadata) GetNamespace() string {
	return m.Namespace
}

func (m *Metadata) GetWorkspace() string {
	return m.Workspace
}

func (m *Metadata) GetTenant() string {
	return m.Tenant
}

func (m *Metadata) GenerateVersion() IObject {
	m.Version = fmt.Sprintf("%d", time.Now().Unix())
	if m.UUID == "" {
		m.UUID = uuid.NewSUID().String()
	}
	return m
}

func Clone(src, tag IObject) {
	b, _ := json.Marshal(src.GetMateData())
	_ = json.Unmarshal(b, tag)
}

func ToMap(i interface{}) (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	bs, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &result); err != nil {
		return nil, err
	}
	return result, err
}

func EncodeFromMap(i IObject, m map[string]interface{}) error {
	bs, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, i); err != nil {
		return err
	}
	return nil
}

type IObject interface {
	GetKind() Kind
	SetKind(Kind)
	GetName() string
	GetNamespace() string
	GetWorkspace() string
	GetTenant() string
	Clone() IObject
	GenerateVersion() IObject
	GetResourceVersion() string
	GetUUID() string
	GetMateData() Metadata
	Delete()
}

func Copy(desc, src IObject) error {
	bs, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, desc); err != nil {
		return err
	}
	return nil
}

func UnmarshalToIObject(data map[string]interface{}, obj IObject) error {
	bs, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, obj); err != nil {
		return err
	}
	return nil
}

type IObjectList interface {
	GenerateListVersion()
}

type Items []IObject

type ObjectList struct {
	Metadata `json:"metadata"`
	Items    `json:"items"`
}

func (iol *ObjectList) GenerateListVersion() {
	var maxVersion string
	for _, item := range iol.Items {
		if item.GetResourceVersion() > maxVersion {
			maxVersion = item.GetResourceVersion()
		}
	}
	iol.Metadata = Metadata{
		Version: maxVersion,
	}
}

func NewIObjectList(items Items) IObjectList {
	iol := &ObjectList{Items: items}
	iol.GenerateListVersion()
	return iol
}

var _ IObject = &DefaultObject{}

type DefaultObject struct {
	Metadata `json:"metadata"`
	Spec     interface{} `json:"spec"`
}

func (i *DefaultObject) GetMateData() Metadata {
	return i.Metadata
}

func (i *DefaultObject) Clone() IObject {
	result := &DefaultObject{}
	Clone(i, result)
	return result
}

func ToItems(objects ...IObject) (result []IObject) {
	result = append(result, objects...)
	return
}

type EventType = string

const (
	ADDED    EventType = "ADDED"
	MODIFIED EventType = "MODIFIED"
	DELETED  EventType = "DELETED"
	REMOVED  EventType = "REMOVED"
)

type Event struct {
	Type   EventType `json:"type"`
	Object IObject   `json:"object"`
}
