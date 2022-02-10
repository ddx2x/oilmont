package iam

import (
	"fmt"
	"time"

	"github.com/ddx2x/oilmont/pkg/service"
	"github.com/ddx2x/oilmont/pkg/service/system"

	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/cache"
	"github.com/ddx2x/oilmont/pkg/micro/webservice"
	iamService "github.com/ddx2x/oilmont/pkg/service/iam"
)

type iamServer struct {
	api.IAPIServer
	webServer webservice.Server

	*iamService.AccountService
	*iamService.BusinessGroupService
	*iamService.RoleService
	*system.OperationService
	*system.ResourceService
	*iamService.UserService
}

func (i *iamServer) Run() error {
	return i.webServer.Run()
}

func NewIAMServer(serviceName string, storage datasource.IStorage) (*iamServer, error) {
	c := cache.NewCache(15*time.Minute, 20*time.Minute)
	baseService := service.NewBaseService(storage, c)
	server := api.NewBaseAPIServer(baseService)

	iam := &iamServer{
		IAPIServer:           server,
		AccountService:       iamService.NewAccount(server.IService),
		BusinessGroupService: iamService.NewBusinessGroup(server.IService),
		RoleService:          iamService.NewRole(server.IService),
		UserService:          iamService.NewUser(server.IService),
	}

	webServer, err := webservice.NewWEBServer(serviceName, "", iam.Server())
	if err != nil {
		return nil, err
	}
	iam.webServer = webServer

	group := iam.Server().Group(fmt.Sprintf("/%s", serviceName))

	// account
	{
		api.GenerateURIV2(group, "iam.ddx2x.nip", "v1", "account", true,
			iam.ListAccount,
			iam.GetAccount,
			iam.CreateAccount,
			iam.UpdateAccount,
			iam.DeleteAccount,
		)
	}

	// businessGroup
	{
		api.GenerateURIV2(group, "iam.ddx2x.nip", "v1", "businessgroup", true,
			iam.ListBusinessGroup,
			iam.GetBusinessGroup,
			iam.CreateBusinessGroup,
			iam.UpdateBusinessGroup,
			iam.DeleteBusinessGroup,
		)
	}

	// role
	{
		api.GenerateURIV2(group, "iam.ddx2x.nip", "v1", "role", true,
			iam.ListRole,
			iam.GetRole,
			iam.CreateRole,
			iam.UpdateRole,
			iam.DeleteRole,
		)
	}
	// user
	{
		api.GenerateURIV2(group, "iam.ddx2x.nip", "v1", "user", true,
			iam.ListUser,
			iam.GetUser,
			nil,
			iam.UpdateUser,
			nil,
		)
	}

	return iam, nil
}
