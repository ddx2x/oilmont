package system

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type ClusterService struct {
	service.IService
}

func NewClusterService(i service.IService) *ClusterService {
	return &ClusterService{i}
}

func (s *ClusterService) List(name string) (core.IObjectList, error) {
	filter := make(map[string]interface{})
	if name != "" {
		filter[common.FilterName] = name
	}
	data := make([]system.Cluster, 0)
	err := s.IService.ListToObject(common.DefaultDatabase, common.CLUSTER, filter, &data, true)
	if err != nil {
		return nil, err
	}

	items := make([]core.IObject, 0)
	for _, item := range data {
		itemP := item
		items = append(items, core.ToItems(&itemP)...)
	}
	return core.NewIObjectList(items), nil
}

func (s *ClusterService) GetByName(name string) (*system.Cluster, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}

	object := &system.Cluster{}
	err := s.IService.GetByFilter(common.DefaultDatabase, common.CLUSTER, object, filter, true)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (s *ClusterService) Update(name string, request *system.Cluster) (core.IObject, bool, error) {
	new, update, err := s.IService.Apply(common.DefaultDatabase, common.CLUSTER, name, request, false)
	if err != nil {
		return nil, false, err
	}
	return new, update, nil
}

func (s *ClusterService) Delete(name string) (core.IObject, error) {
	object, err := s.GetByName(name)
	if err != nil {
		return nil, err
	}
	err = s.DeleteObject(common.DefaultDatabase, common.CLUSTER, name, object, true)
	return object, err
}
