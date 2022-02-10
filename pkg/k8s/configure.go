package k8s

import (
	"context"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"kubevirt.io/client-go/kubecli"
)

type Configure struct {
	ResourceRegistry
	// kubernetes reset config
	*rest.Config
	// k8s DynamicSharedInformerFactory
	dynamicinformer.DynamicSharedInformerFactory
	// k8s dyc client
	dynamic.Interface
	//Clientset
	*kubernetes.Clientset
	// kubernetes discovery client
	*discovery.DiscoveryClient
	// kubevirt client
	kubecli.KubevirtClient
	// ?
	Stop func()
}

func NewKubeConfigureFromJSONCfg(ctx context.Context, rif ResourceRegistry, configV1 *clientcmdapiv1.Config) (*Configure, error) {
	var restCfg *rest.Config
	var err error

	if restCfg, err = buildRestCfgFromJSON(configV1); err != nil {
		return nil, err
	}
	return createConfigure(ctx, restCfg, rif)
}

func NewKubeConfigure(ctx context.Context, rif ResourceRegistry, kubeConfig string, inCluster bool) (*Configure, error) {
	var restCfg *rest.Config
	var err error
	if inCluster {
		restCfg, err = createCfgFromCluster()
	} else {
		restCfg, err = createCfgFromPath(kubeConfig)
	}

	if err != nil {
		return nil, err
	}
	return createConfigure(ctx, restCfg, rif)
}

func createConfigure(ctx context.Context, restCfg *rest.Config, rif ResourceRegistry) (*Configure, error) {
	var dynInt dynamic.Interface
	var clientset *kubernetes.Clientset
	var err error
	stopC := make(chan struct{}, rif.Length())
	stop := func() {
		<-ctx.Done()
		for range iter(rif.Length()) {
			stopC <- struct{}{}
		}
	}
	clientset, dynInt, err = buildClientSet(restCfg)
	if err != nil {
		return nil, err
	}

	dynamicSharedInformerFactory, err := createInformerFactory(rif, dynInt, stopC)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return nil, err
	}

	kubevirtCli, err := kubecli.GetKubevirtClientFromRESTConfig(restCfg)
	if err != nil {
		return nil, err
	}

	return &Configure{
		Stop:             stop,
		Interface:        dynInt,
		Config:           restCfg,
		Clientset:        clientset,
		DiscoveryClient:  discoveryClient,
		ResourceRegistry: rif,
		KubevirtClient:   kubevirtCli,

		DynamicSharedInformerFactory: dynamicSharedInformerFactory,
	}, nil
}

func iter(n int) []struct{} { return make([]struct{}, n) }
