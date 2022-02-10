package controller

import (
	"context"
	"encoding/json"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	objUtils "github.com/ddx2x/oilmont/pkg/utils/obj"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type Controller interface {
	Run() error
}

type NorthHandler interface {
	NorthOnAdd(obj core.IObject)
	NorthOnUpdate(obj core.IObject)
	NorthOnDelete(obj core.IObject)
	NorthEventCh(ctx context.Context) (<-chan core.Event, error)
}

type SouthHandler interface {
	SouthOnAdd(obj runtime.Object)
	SouthOnUpdate(obj runtime.Object)
	SouthOnDelete(obj runtime.Object)
	SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error)
}

type InjectClient interface {
	Set(cs *clients.Clients, stage datasource.IStorage)
}

type Handler interface {
	SouthHandler
	NorthHandler
	InjectClient
}

type Controllers struct {
	stage              datasource.IStorage
	clients            *clients.Clients
	backendControllers []*BackendController
}

func NewControllers(stage datasource.IStorage) *Controllers {
	cs := clients.NewClients()
	return &Controllers{
		stage:              stage,
		clients:            cs,
		backendControllers: make([]*BackendController, 0),
	}
}

func (c *Controllers) Add(handlers ...Handler) error {
	for _, h := range handlers {
		bc, err := NewBackendController(c.stage, c.clients, h)
		if err != nil {
			return err
		}
		c.backendControllers = append(c.backendControllers, bc)
	}

	return nil
}

func (c *Controllers) Run(ctx context.Context) chan error {
	errChan := make(chan error, 0)

	clusterCh, err := c.stage.WatchEvent(ctx, common.DefaultDatabase, common.CLUSTER, "0")
	if err != nil {
		errChan <- err
		return errChan
	}

	if err := c.SetupCluster(ctx, clusterCh); err != nil {
		errChan <- err
		return errChan
	}

	for _, controller := range c.backendControllers {
		//controller.Set(c.clients, c.stage)
		go controller.Start(ctx, errChan)
	}
	return errChan
}

type BackendController struct {
	stage   datasource.IStorage
	clients *clients.Clients
	Handler
}

func NewBackendController(stage datasource.IStorage, cs *clients.Clients, h Handler) (*BackendController, error) {
	//cs := clients.NewClients()
	h.Set(cs, stage)
	return &BackendController{
		stage:   stage,
		clients: cs,
		Handler: h,
	}, nil
}

func (bc *BackendController) northHandle(event core.Event) {
	switch event.Type {
	case core.ADDED:
		bc.NorthOnAdd(event.Object)
	case core.MODIFIED:
		bc.NorthOnUpdate(event.Object)
	case core.DELETED:
		bc.NorthOnDelete(event.Object)
	}
}
func (bc *BackendController) southHandle(event watch.Event) {
	switch event.Type {
	case watch.Added:
		bc.SouthOnAdd(event.Object)
	case watch.Modified:
		bc.SouthOnUpdate(event.Object)
	case watch.Deleted:
		bc.SouthOnDelete(event.Object)
	}
}

func (c *Controllers) SetupCluster(ctx context.Context, ec <-chan core.Event) error {
	clusters := make([]system.Cluster, 0)
	data, err := c.stage.List(common.DefaultDatabase, common.CLUSTER, "", true)

	if err != nil {
		return err
	}
	err = objUtils.UnstructuredObjectToInstanceObj(data, &clusters)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		bs, err := json.Marshal(cluster.Spec.Config)
		if err != nil {
			log.G(ctx).Warnf("backend controller read cluster config marshal error:%s err")
		}

		cfg, err := clients.JSONToConfig(bs)
		if err != nil {
			log.G(ctx).Warnf("backend controller create kubeclient error:%s err")
			continue
		}
		kubecli, err := clients.BuildClient(cluster.GetName(), *cfg)
		if err != nil {
			log.G(ctx).Warnf("backend controller create build client error:%s err")
			continue
		}
		c.clients.AddKubeClient(cluster.GetName(), kubecli)
	}
	go func() {
		for {
			select {
			case e, ok := <-ec:
				if !ok {
					return
				}
				cluster := e.Object.(*core.DefaultObject)
				switch e.Type {
				case core.ADDED, core.MODIFIED:
					bs, err := json.Marshal(cluster.Spec)
					if err != nil {
						log.G(ctx).Warnf("backend controller read cluster config marshal error:%s err")
					}

					data := map[string]interface{}{}
					err = json.Unmarshal(bs, &data)
					if err != nil {
						log.G(ctx).Warnf("backend controller read cluster config marshal error:%s err")
					}
					bs, err = json.Marshal(data["config"])

					cfg, err := clients.JSONToConfig(bs)
					if err != nil {
						log.G(ctx).Warnf("backend controller create kubeclient error:%s err")
						continue
					}
					kubecli, err := clients.BuildClient(cluster.GetName(), *cfg)
					if err != nil {
						log.G(ctx).Warnf("backend controller create build client error:%s err")
						continue
					}
					c.clients.AddKubeClient(cluster.GetName(), kubecli)
				case core.DELETED:
					c.clients.RemoveClient(cluster.GetName())
				}
			}
		}
	}()

	return nil
}

func (bc *BackendController) Start(ctx context.Context, errChan chan error) {

	northEvent, err := bc.NorthEventCh(ctx)
	if err != nil {
		errChan <- err
		return
	}
	southEvents, err := bc.SouthEventChs(ctx)
	if err != nil {
		errChan <- err
		return
	}

	channels := len(southEvents) + 1
	stopCh := make(chan struct{}, channels)

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case event, ok := <-northEvent:
				if !ok {
					return
				}
				bc.northHandle(event)
			}
		}
	}()

	go func() {
		for _, southEvent := range southEvents {
			go func(southEvent <-chan watch.Event) {
				for {
					select {
					case <-stopCh:
						return
					case event, ok := <-southEvent:
						if !ok {
							return
						}
						bc.southHandle(event)
					}
				}
			}(southEvent)
		}
	}()

	go func() {
		<-ctx.Done()
		for range iter(channels) {
			stopCh <- struct{}{}
		}
	}()

	return
}

func iter(n int) []struct{} { return make([]struct{}, n) }
