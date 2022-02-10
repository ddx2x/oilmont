package event

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"github.com/ddx2x/oilmont/pkg/service"
)

type CloudEventService struct {
	service.IService
}

func NewCloudEvent(i service.IService) *CloudEventService {
	return &CloudEventService{i}
}

func (ce *CloudEventService) List(name, namespace string) (*event.CloudEventList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	if namespace != "" {
		filter[common.FilterNamespace] = namespace
	}

	data := make([]event.CloudEvent, 0)
	err := ce.IService.ListToObject(common.DefaultDatabase, common.CLOUDEVENT, filter, &data, true)
	if err != nil {
		return nil, err
	}

	objList := &event.CloudEventList{Items: data}
	objList.GenerateListVersion()

	return objList, nil
}

func (ce *CloudEventService) GetByName(namespace, name string) (*event.CloudEvent, error) {
	filter := map[string]interface{}{
		common.FilterName:      name,
		common.FilterNamespace: namespace,
	}

	respObj := &event.CloudEvent{}
	err := ce.IService.GetByFilter(common.DefaultDatabase, common.CLOUDEVENT, respObj, filter, true)
	if err != nil {
		return nil, err
	}
	return respObj, nil
}
