package kes

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"
)

func init() {
	k8s.ShardingResourceRegistry.Register("pods", k8s.GVR{Group: "", Version: "v1", Resource: "pods"})
}

func (k *KesServer) listPods(g *gin.Context) {
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

	list, err := cli.Resource(k8s.GVR{Group: "", Version: "v1", Resource: "pods"}).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.RequestParametersError(g, err)
		return
	}

	SetUnstructuredListBaseCloudInfo("lizhi", "local", "local-1", "defualt", list)
	g.JSONP(http.StatusOK, list)
}

func (k *KesServer) DeletePod(g *gin.Context) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	namespace := g.Param("namespace")
	name := g.Param("name")
	if namespace == "" || name == "" {
		api.RequestParametersError(g, fmt.Errorf("params not obtain namespace=%s name=%s", namespace, name))
		return
	}

	cluster := g.Query("cluster")
	if cluster == "" {
		cluster = "default"
	}

	cli := k.multiCluster.Get(cluster)
	if cli == nil {
		api.RequestParametersError(g, fmt.Errorf("request cluster %s is not found", cluster))
		return
	}

	gvr, err := k8s.ShardingResourceRegistry.GetGVR("pods")
	if err != nil {
		api.RequestParametersError(g, fmt.Errorf("request gvr %s is not found", cluster))
		return
	}

	err = cli.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		api.RequestParametersError(g, fmt.Errorf("delete err %s", err))
		return
	}

	g.JSON(http.StatusOK, "")
}

type logRequest struct {
	Container  string    `form:"container" json:"container"`
	Timestamps bool      `form:"timestamps" json:"timestamps"`
	SinceTime  time.Time `form:"sinceTime" json:"sinceTime"`
	TailLines  int64     `form:"tailLines" json:"tailLines"`
}

func (k *KesServer) LogsPod(g *gin.Context) {
	namespace := g.Param("namespace")
	name := g.Param("name")

	lq := &logRequest{}
	if err := g.Bind(lq); err != nil || namespace == "" || name == "" {
		api.RequestParametersError(g, fmt.Errorf("params not obtain namespace=%s name=%s", namespace, name))
		return
	}

	cluster := g.Query("cluster")
	if cluster == "" {
		cluster = "default"
	}

	cli := k.multiCluster.Get(cluster)
	if cli == nil {
		api.RequestParametersError(g, fmt.Errorf("request cluster %s is not found", cluster))
		return
	}

	buf := bytes.NewBufferString("")
	err := k.Logs(
		cli,
		namespace,
		name,
		lq.Container,
		false,
		false,
		lq.Timestamps,
		0,
		&lq.SinceTime,
		0,
		lq.TailLines,
		buf,
	)

	if err != nil {
		api.InternalServerError(g, err, err)
		return
	}

	g.JSON(http.StatusOK, buf.String())
}
