package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type WorkspaceService struct {
	service.IService
}

func NewWorkspaceService(i service.IService) *WorkspaceService {
	return &WorkspaceService{i}
}

func (ws *WorkspaceService) List(name string) (*system.WorkspaceList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}

	data := make([]system.Workspace, 0)
	err := ws.IService.ListToObject(common.DefaultDatabase, common.WORKSPACE, filter, &data, true)
	if err != nil {
		return nil, err
	}

	workspaceList := &system.WorkspaceList{Items: data}
	workspaceList.GenerateListVersion()

	return workspaceList, nil
}

func (ws *WorkspaceService) GetByName(name string) (*system.Workspace, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}
	workspaceObj := &system.Workspace{}
	err := ws.IService.GetByFilter(common.DefaultDatabase, common.WORKSPACE, workspaceObj, filter, true)
	if err != nil {
		return nil, err
	}

	return workspaceObj, nil
}

func (ws *WorkspaceService) Create(reqWorkspace *system.Workspace) (core.IObject, error) {
	if _, err := ws.GetByName(reqWorkspace.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("workspace name %s exists", reqWorkspace.Name)
	}
	if reqWorkspace.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if reqWorkspace.Spec.Tenant == "" {
		return nil, fmt.Errorf("tenant is empty")
	}

	reqWorkspace.Workspace = reqWorkspace.Spec.Tenant
	new, err := ws.IService.Create(common.DefaultDatabase, common.WORKSPACE, reqWorkspace)
	if err != nil {
		return nil, err
	}
	return new, nil
}

func (ws *WorkspaceService) Update(name string, request *system.Workspace, paths ...string) (core.IObject, bool, error) {
	new, update, err := ws.IService.Apply(common.DefaultDatabase, common.WORKSPACE, name, request, false, paths...)
	if err != nil {
		return nil, false, err
	}
	return new, update, nil
}

func (ws *WorkspaceService) Delete(name string) (core.IObject, error) {
	object, err := ws.GetByName(name)
	if err != nil {
		return nil, err
	}

	err = ws.DeleteObject(common.DefaultDatabase, common.WORKSPACE, name, object, true)
	return object, err
}
