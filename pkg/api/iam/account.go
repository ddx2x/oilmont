package iam

import (
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
)

func (i *iamServer) ListAccount(g *gin.Context) {
	namespace := g.Param("namespace")
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	results, err := i.AccountService.List(tenant, namespace)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}
	g.JSON(http.StatusOK, results)
}

func (i *iamServer) GetAccount(g *gin.Context) {
	name := g.Param("name")
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	results, err := i.AccountService.List(tenant, name)
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSON(http.StatusOK, results)
}

func (i *iamServer) CreateAccount(g *gin.Context) {
	request := &iam.Account{}
	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	// permission check.

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	//if !i.AccountService.CheckCreateAccountWorkspacePermission(reqUser, request.Workspace) ||
	//	!i.AccountService.CheckCreateAccountTypePermission(reqUser, request.Spec.AccountType) {
	//	i.RecordEvent(common.ACCOUNT, core.ADDED, reqUser, request, core.CloudEventFail)
	//	api.RequestParametersError(g, fmt.Errorf("account create reject"))
	//	return
	//}

	res, err := i.AccountService.Create(request)
	if err != nil {
		i.RecordEvent(common.ACCOUNT, core.ADDED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.ACCOUNT, core.ADDED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *iamServer) UpdateAccount(g *gin.Context) {
	name := g.Param("name")
	paths := api.SplitPath(g.DefaultQuery("path", ""))
	tenant := g.Query("tenant")
	request := &iam.Account{}

	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)

	if err := g.ShouldBindJSON(request); err != nil {
		api.RequestParametersError(g, err)
		return
	}

	res, _, err := i.AccountService.Update(tenant, "", name, request, paths...)
	if err != nil {
		i.RecordEvent(common.ACCOUNT, core.MODIFIED, reqUser, request, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.ACCOUNT, core.MODIFIED, reqUser, res, event.CloudEventSuccess)
	g.JSON(http.StatusOK, res)
}

func (i *iamServer) DeleteAccount(g *gin.Context) {
	name := g.Param("name")
	namespace := g.Param("namespace")
	reqUser := g.GetHeader(common.HttpRequestUserHeaderKey)
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	account, err := i.AccountService.Delete(tenant, namespace, name)
	if err != nil {
		i.RecordEvent(common.ACCOUNT, core.DELETED, reqUser, account, event.CloudEventFail)
		api.RequestParametersError(g, err)
		return
	}

	i.RecordEvent(common.ACCOUNT, core.DELETED, reqUser, account, event.CloudEventSuccess)
	g.JSON(http.StatusOK, account)
}
