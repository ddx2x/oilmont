package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/resource/system"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/service"
)

type OperationService struct {
	service.IService
}

func NewOperation(i service.IService) *OperationService {
	return &OperationService{i}
}

func (o *OperationService) List(name string) (*system.OperationList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}

	operations := make([]system.Operation, 0)
	err := o.IService.ListToObject(common.DefaultDatabase, common.OPERATION, filter, &operations, true)
	if err != nil {
		return nil, err
	}

	operationList := &system.OperationList{Items: operations}
	operationList.GenerateListVersion()

	return operationList, nil
}

func (o *OperationService) GetByName(name string) (*system.Operation, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}

	operation := &system.Operation{}

	if err := o.IService.GetByFilter(common.DefaultDatabase, common.OPERATION, operation, filter, true); err != nil {
		return nil, err
	}
	return operation, nil
}

func (o *OperationService) Create(reqOperation *system.Operation) (*system.Operation, error) {
	if _, err := o.GetByName(reqOperation.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("operation exists")
	}

	if reqOperation.Spec.Method == "" || reqOperation.Spec.OP == "" {
		return nil, fmt.Errorf("operation params is empty")
	}

	reqOperation.Kind = system.OperationKind
	reqOperation.GenerateVersion()

	_, err := o.IService.Create(common.DefaultDatabase, common.OPERATION, reqOperation)
	if err != nil {
		return nil, err
	}
	return reqOperation, nil
}

func (o *OperationService) Update(name string, reqOperation *system.Operation) (core.IObject, bool, error) {
	operation, err := o.GetByName(name)
	if err != nil {
		return nil, false, err
	}

	operation.Spec.Method = reqOperation.Spec.Method
	operation.Spec.OP = reqOperation.Spec.OP

	if operation.Spec.Method == "" || operation.Spec.OP == "" {
		return nil, false, fmt.Errorf("operation params is empty")
	}

	_, update, err := o.IService.Apply(common.DefaultDatabase, common.OPERATION, operation.Name, operation, false)
	if err != nil {
		return nil, false, err
	}
	return operation, update, nil
}

func (o *OperationService) Delete(name string) (core.IObject, error) {
	object, err := o.GetByName(name)
	if err != nil {
		return nil, err
	}
	err = o.DeleteObject(common.DefaultDatabase, common.OPERATION, object.GetName(), object, true)
	return object, err
}
