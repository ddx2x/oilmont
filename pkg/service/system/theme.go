package system

import (
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	"github.com/ddx2x/oilmont/pkg/service"
)

type ThemeService struct {
	service.IService
}

func NewThemeService(i service.IService) *ThemeService {
	return &ThemeService{i}
}

func (ts *ThemeService) List(name string) (*system.ThemeList, error) {
	filter := map[string]interface{}{}
	if name != "" {
		filter[common.FilterName] = name
	}

	data := make([]system.Theme, 0)
	err := ts.IService.ListToObject(common.DefaultDatabase, common.THEME, filter, &data, true)
	if err != nil {
		return nil, err
	}

	themeList := &system.ThemeList{Items: data}
	themeList.GenerateListVersion()

	return themeList, nil
}

func (ts *ThemeService) GetByName(name string) (*system.Theme, error) {
	filter := map[string]interface{}{
		common.FilterName: name,
	}

	obj := &system.Theme{}
	err := ts.IService.GetByFilter(common.DefaultDatabase, common.THEME, obj, filter, true)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (ts *ThemeService) Create(reqObj *system.Theme) (core.IObject, error) {
	if _, err := ts.GetByName(reqObj.Name); err != datasource.NotFound {
		return nil, fmt.Errorf("theme exists")
	}

	if reqObj.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	reqObj.Kind = system.ThemeKind
	reqObj.GenerateVersion()

	_, err := ts.IService.Create(common.DefaultDatabase, common.THEME, reqObj)
	if err != nil {
		return nil, err
	}
	return reqObj, nil
}

func (ts *ThemeService) Update(name string, reqObj *system.Theme) (core.IObject, bool, error) {
	obj, err := ts.GetByName(name)
	if err != nil {
		return nil, false, err
	}

	obj.Spec = reqObj.Spec

	_, update, err := ts.IService.Apply(common.DefaultDatabase, common.THEME, obj.Name, obj, false)
	if err != nil {
		return nil, false, err
	}

	return obj, update, nil
}

func (ts *ThemeService) Delete(name string) (core.IObject, error) {
	object, err := ts.GetByName(name)
	if err != nil {
		return nil, err
	}
	err = ts.DeleteObject(common.DefaultDatabase, common.THEME, name, object, true)

	return object, err
}
