package common

import (
	"flag"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

var (
	// InCluster Flag for the application runtime
	InCluster bool
	// KubeConfig DefaultConfigFile is the default bootstrap configuration
	KubeConfig *string
)

func init() {
	if v := os.Getenv("IN_CLUSTER"); v != "" {
		InCluster = true
	}
	if v := os.Getenv("KUBE_CONFIG"); v != "" && !InCluster {
		*KubeConfig = v
	}
	if home := homedir.HomeDir(); home != "" {
		KubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		KubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}
