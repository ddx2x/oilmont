package k8s

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	k8sjson "k8s.io/apimachinery/pkg/util/json"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sync"
)

type MultiCluster struct {
	clusters map[string]*Configure
	mutex    *sync.Mutex
	storage  datasource.IStorage
}

func NewMultiCluster(storage datasource.IStorage) *MultiCluster {
	return &MultiCluster{
		clusters: make(map[string]*Configure),
		mutex:    &sync.Mutex{},
		storage:  storage,
	}
}

func (m *MultiCluster) set(name string, cfg *Configure) {
	m.mutex.Lock()
	m.clusters[name] = cfg
	m.mutex.Unlock()
}

func (m *MultiCluster) remove(name string) {
	m.mutex.Lock()
	if cli := m.clusters[name]; cli != nil {
		cli.Stop()
	}
	delete(m.clusters, name)
	m.mutex.Unlock()
}

func (m *MultiCluster) Get(name string) *Configure {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.clusters[name]
}

func (m *MultiCluster) AsyncRun(ctx context.Context) error {
	flog := log.G(ctx)
	apply := func(cluster *system.Cluster) error {
		cmdcfg := &clientcmdapiv1.Config{}
		src, err := k8sjson.Marshal(cluster.Spec.Config)
		if err != nil {
			return err
		}
		if err := k8sjson.Unmarshal(src, cmdcfg); err != nil {
			return err
		}
		cfg, err := NewKubeConfigureFromJSONCfg(ctx, ShardingResourceRegistry, cmdcfg)
		if err != nil {
			return err
		}
		m.set(cluster.GetName(), cfg)
		return nil
	}

	clusters := make([]system.Cluster, 0)
	if err := m.storage.ListToObject(common.DefaultDatabase, common.CLUSTER, nil, &clusters, true); err != nil {
		return err
	}

	for _, cluster := range clusters {
		c := cluster
		if err := apply(&c); err != nil {
			flog.Warnf("init cluster %s config error: %s", cluster.GetName(), err)
		}
	}

	clusterList := &system.ClusterList{Items: clusters}
	clusterList.GenerateVersion()

	eventCh, err := m.storage.WatchEvent(ctx, common.DefaultDatabase, common.CLUSTER, clusterList.GetResourceVersion())
	if err != nil {
		return err
	}

	onEvent := func(event core.Event) error {
		flog.Infof("onEvent handle cluster op: %s cluster: %s", event.Type, event.Object.GetName())
		switch event.Type {
		case core.ADDED, core.MODIFIED:
			cluster := &system.Cluster{}
			if err := core.Copy(cluster, event.Object); err != nil {
				flog.Warnf("can't copy object to cluster error: %s", err)
				return err
			}
			if err := apply(cluster); err != nil {
				flog.Warnf("can't apply to clusters error: %s", err)
				return err
			}
		case core.DELETED:
			m.remove(event.Object.GetName())
		}
		return nil
	}

	go func(eventCh <-chan core.Event) {
		for {
			select {
			case event, ok := <-eventCh:
				if !ok {
					return
				}
				if err := onEvent(event); err != nil {
					continue
				}
			}
		}
	}(eventCh)
	return nil
}
