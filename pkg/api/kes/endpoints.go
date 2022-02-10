package kes

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func init() {
	k8s.ShardingResourceRegistry.Register("endpoints", k8s.GVR{Group: "", Version: "v1", Resource: "endpoints"})
}

func (k *KesServer) listEndpoints(g *gin.Context) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	cluster := g.Query("cluster")
	if cluster == "" {
		cluster = "default"
	}

	cli := k.multiCluster.Get(cluster)
	if cli == nil {
		api.RequestParametersError(g, fmt.Errorf("request cluster %s is not found", cluster))
		return
	}

	gvr, err := k8s.ShardingResourceRegistry.GetGVR("endpoints")
	if err != nil {
		api.RequestParametersError(g, fmt.Errorf("request gvr %s is not found", cluster))
		return
	}

	list, err := cli.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	g.JSONP(http.StatusOK, list)
}
