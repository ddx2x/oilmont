package k8s

import (
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"time"

	"k8s.io/client-go/dynamic"
	client "k8s.io/client-go/dynamic"
	informers "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// full sync cache resource time
	period = 30 * time.Second
)

func buildRestCfgFromJSON(configV1 *clientcmdapiv1.Config) (*rest.Config, error) {
	configObject, err := clientcmdlatest.Scheme.ConvertToVersion(configV1, clientcmdapi.SchemeGroupVersion)
	configInternal := configObject.(*clientcmdapi.Config)

	cfg, err := clientcmd.NewDefaultClientConfig(
		*configInternal,
		&clientcmd.ConfigOverrides{
			ClusterDefaults: clientcmdapi.Cluster{
				Server: "",
			},
		},
	).ClientConfig()

	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func buildClientSet(cfg *rest.Config) (*kubernetes.Clientset, client.Interface, error) {
	_interface, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	return clientset, _interface, nil
}

func createInformerFactory(rif ResourceRegistry, _interface dynamic.Interface, stopC <-chan struct{}) (informers.DynamicSharedInformerFactory, error) {
	dsif := informers.NewDynamicSharedInformerFactory(_interface, period)
	rif.Subscript(dsif, stopC)
	return dsif, nil
}

func createCfgFromPath(path string) (*rest.Config, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	cfg = applyDefaultRateLimiter(cfg, 2)
	return cfg, nil
}

func createCfgFromCluster() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	cfg = applyDefaultRateLimiter(cfg, 2)
	return cfg, nil
}

func applyDefaultRateLimiter(config *rest.Config, flowRate int) *rest.Config {
	if flowRate < 0 {
		flowRate = 1
	}
	// here we magnify the default qps and burst in client-go
	config.QPS = rest.DefaultQPS * float32(flowRate)
	config.Burst = rest.DefaultBurst * flowRate

	return config
}
