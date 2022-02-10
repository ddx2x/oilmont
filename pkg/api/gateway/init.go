package gateway

import "github.com/ddx2x/oilmont/pkg/k8s"

func init() {
	k8s.ShardingResourceRegistry.Register("pods", k8s.GVR{Group: "", Version: "v1", Resource: "pods"})
}
