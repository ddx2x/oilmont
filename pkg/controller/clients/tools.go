package clients

import (
	"github.com/ddx2x/oilmont/pkg/datasource/k8s"
	k8sjson "k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"kubevirt.io/client-go/kubecli"
)

const (
	// High enough QPS to fit all expected use cases.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases.
	defaultBurst = 1e6
	// full resyc cache resource time
	//defaultResyncPeriod = 30 * time.Second
)

type K8sConfig struct {
	Name   string                `json:"name"`
	Config clientcmdapiv1.Config `json:"config"`
}

func JSONToConfig(data []byte) (*clientcmdapiv1.Config, error) {
	config := &clientcmdapiv1.Config{}
	err := k8sjson.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func BuildClient(master string, configV1 clientcmdapiv1.Config) (*KubeClient, error) {
	configObject, err := clientcmdlatest.Scheme.ConvertToVersion(&configV1, clientcmdapi.SchemeGroupVersion)
	configInternal := configObject.(*clientcmdapi.Config)

	cfg, err := clientcmd.NewDefaultClientConfig(
		*configInternal,
		&clientcmd.ConfigOverrides{
			ClusterDefaults: clientcmdapi.Cluster{
				Server: master,
			},
		},
	).ClientConfig()

	if err != nil {
		return nil, err
	}

	cfg.QPS = defaultQPS
	cfg.Burst = defaultBurst

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	_interface, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	kubevirtCli, err := kubecli.GetKubevirtClientFromRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	drs := k8s.NewIDataSource(_interface)

	return &KubeClient{
		ClientSet:   clientSet,
		Cfg:         cfg,
		Interface:   _interface,
		Service:     drs,
		KubevirtCli: kubevirtCli,
	}, nil
}
