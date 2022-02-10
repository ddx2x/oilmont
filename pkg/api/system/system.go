package system

import (
	"fmt"
	"time"

	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/cache"
	"github.com/ddx2x/oilmont/pkg/micro/webservice"
	"github.com/ddx2x/oilmont/pkg/service"
	"github.com/ddx2x/oilmont/pkg/service/system"
)

type systemServer struct {
	api.IAPIServer
	webServer        webservice.Server
	menu             *system.MenuService
	cluster          *system.ClusterService
	availableZone    *system.AvailableZoneService
	region           *system.RegionService
	workspace        *system.WorkspaceService
	license          *system.LicenseService
	provider         *system.ProviderService
	instanceType     *system.InstanceTypeService
	operationService *system.OperationService
	resourceService  *system.ResourceService
	themeService     *system.ThemeService
	tenant           *system.TenantService
}

func (i *systemServer) Run() error {
	return i.webServer.Run()
}

func NewSystemServer(serviceName string, storage datasource.IStorage) (*systemServer, error) {
	c := cache.NewCache(15*time.Minute, 20*time.Minute)
	baseService := service.NewBaseService(storage, c)
	baseServer := api.NewBaseAPIServer(baseService)

	server := &systemServer{
		IAPIServer:       baseServer,
		cluster:          system.NewClusterService(baseService),
		menu:             system.NewMenuService(baseService),
		availableZone:    system.NewAvailableZoneService(baseService),
		region:           system.NewRegionService(baseService),
		workspace:        system.NewWorkspaceService(baseService),
		license:          system.NewLicenseService(baseService),
		provider:         system.NewProviderService(baseService),
		instanceType:     system.NewInstanceType(baseService),
		operationService: system.NewOperation(baseService),
		resourceService:  system.NewResource(baseService),
		themeService:     system.NewThemeService(baseService),
		tenant:           system.NewTenant(baseService),
	}

	webServer, err := webservice.NewWEBServer(serviceName, "", server.Server())
	if err != nil {
		return nil, err
	}
	server.webServer = webServer

	group := server.Server().Group(fmt.Sprintf("/%s", serviceName))

	// cluster
	{

		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "cluster", false,
			server.ListCluster,
			server.ListCluster,
			server.CreateCluster,
			server.UpdateCluster,
			server.DeleteCluster,
		)
	}

	// availableZone
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "availablezone", false,
			server.ListAvailableZone,
			server.ListAvailableZone,
			server.CreateAvailableZone,
			server.UpdateAvailableZone,
			server.DeleteAvailableZone,
		)
	}

	// region
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "region", true,
			server.ListRegion,
			server.ListRegion,
			server.CreateRegion,
			server.UpdateRegion,
			server.DeleteRegion,
		)
	}

	// workspace
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "workspace", false,
			server.ListWorkspace,
			server.ListWorkspace,
			server.CreateWorkspace,
			server.UpdateWorkspace,
			server.DeleteWorkspace,
		)
	}

	// license
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "license", true,
			server.ListLicense,
			server.ListLicense,
			server.CreateLicense,
			server.UpdateLicense,
			server.DeleteLicense,
		)
	}

	// provider
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "provider", false,
			server.Provider,
			server.Provider,
			server.CreateProvider,
			server.UpdateProvider,
			server.DeleteProvider,
		)
	}

	// menu
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "menu", false,
			server.ListMenu,
			nil,
			server.CreateMenu,
			server.UpdateMenu,
			server.DeleteMenu,
		)
	}

	// instancetype
	{
		group.GET(api.GenerateURI("system.ddx2x.nip", "v1", "instancetype", true), server.ListInstanceType)
		group.GET(api.GenerateURI("system.ddx2x.nip", "v1", "instancetype", false), server.ListInstanceType)
	}

	// operation
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "operation", false,
			server.ListOperation,
			server.GetOperation,
			server.CreateOperation,
			server.UpdateOperation,
			server.DeleteOperation,
		)
	}

	// resource
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "resource", false,
			server.ListResource,
			server.GetResource,
			server.CreateResource,
			server.UpdateResource,
			server.DeleteResource,
		)
	}

	// theme
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "theme", false,
			server.ListTheme,
			server.ListTheme,
			server.CreateTheme,
			server.UpdateTheme,
			server.DeleteTheme,
		)
	}

	// tenant
	{
		api.GenerateURIV2(group, "system.ddx2x.nip", "v1", "tenant", false,
			server.ListTenant,
			server.ListTenant,
			server.CreateTenant,
			server.UpdateTenant,
			server.DeleteTenant,
		)
	}

	return server, nil
}
