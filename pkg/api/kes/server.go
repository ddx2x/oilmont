package kes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/datasource/cache"
	"github.com/ddx2x/oilmont/pkg/k8s"
	"github.com/ddx2x/oilmont/pkg/micro/webservice"
	"github.com/ddx2x/oilmont/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/igm/sockjs-go/v3/sockjs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/remotecommand"
)

type KesServer struct {
	*api.BaseAPIServer
	webServer    webservice.Server
	multiCluster *k8s.MultiCluster
}

func (k *KesServer) Run() error {
	return k.webServer.Run()
}

func NewKesServer(serverName string, storage datasource.IStorage) (*KesServer, error) {
	kes := &KesServer{
		// default extend
		BaseAPIServer: api.NewBaseAPIServer(
			service.NewBaseService(storage, cache.NewCache(15*time.Minute, 20*time.Minute)),
		),
		multiCluster: k8s.NewMultiCluster(storage),
	}

	// register kes server to micro service
	web, err := webservice.NewWEBServer(serverName, "", kes)
	if err != nil {
		return nil, err
	}
	// injector webserver to kes register
	kes.webServer = web

	ctx := context.Background()

	group := kes.Group(common.MicroServiceName(serverName))
	group.GET("/", func(g *gin.Context) {
		g.JSON(http.StatusOK, "")
	})

	// pods
	{
		group.GET("/api/v1/pods", kes.listPods)
		group.GET("/api/v1/namespaces/:namespace/pods/:name/log", kes.LogsPod)
		group.DELETE("/api/v1/namespaces/:namespace/pods/:name", kes.DeletePod)
	}

	// deployments
	{
		group.GET("/api/v1/deployments", kes.listDeployment)
	}

	// statefulSets
	{
		group.GET("/api/v1/statefulsets", kes.listStatefulSets)
	}

	// configMaps
	{
		group.GET("/api/v1/configmaps", kes.listConfigMaps)
	}

	// services
	{
		group.GET("/api/v1/services", kes.listServices)
	}

	// endpoints
	{
		group.GET("/api/v1/endpoints", kes.listEndpoints)
	}

	// ingresses
	{
		group.GET("/api/v1/ingresses", kes.listIngresses)
	}

	// persistentvolumeclaims
	{
		group.GET("/api/v1/persistentvolumeclaims", kes.listPersistentVolumeClaims)
	}

	// persistentvolume
	{
		group.GET("/api/v1/persistentvolumes", kes.listPersistentVolume)
	}

	// storageclasses
	{
		group.GET("/api/v1/storageclasses", kes.listStorageClasses)
	}

	// webshell
	{
		createGlobalSessionManager(kes.multiCluster)
		options := sockjs.DefaultOptions
		options.JSessionID = sockjs.DefaultJSessionID

		httpHandle := sockjs.NewHandler(
			fmt.Sprintf("/%s/shell/pod", serverName),
			options,
			wrapSockjsHandle(ctx),
		)
		group.Any("/shell/pod/*path", gin.WrapH(httpHandle))
		group.GET("/shell/namespace/:namespace/pod/:name/container/:container/:shelltype/:cluster", kes.podAttach)
	}

	//metrics
	{
		group.POST("/metrics", kes.Metrics)
		//group.GET("/apis/metrics.k8s.io/v1beta1/nodes", kes)
		//group.GET("/apis/metrics.k8s.io/v1beta1/pods", workloadServer.ListPodMetrics)
		//group.GET("/apis/metrics.k8s.io/v1beta1/namespaces/:namespace/pods", workloadServer.GetPodMetrics)
	}

	if err := kes.multiCluster.AsyncRun(ctx); err != nil {
		return nil, err
	}

	return kes, nil
}

func (k *KesServer) podAttach(g *gin.Context) {
	attachPodRequest := &attachPodRequest{
		Namespace: g.Param("namespace"),
		Name:      g.Param("name"),
		Container: g.Param("container"),
		ShellType: g.Param("shelltype"),
		Shell:     g.Query("shell"),
		Image:     g.Query("image"),
		Cluster:   g.Query("cluster"),
	}

	sessId, err := generateTerminalSessionId()
	if err != nil {
		api.InternalServerError(g, err, err)
		return
	}

	shardingManager.set(
		sessId,
		&sessionChannel{
			id:       sessId,
			bound:    make(chan struct{}),
			sizeChan: make(chan remotecommand.TerminalSize),
		})

	go waitForTerminal(attachPodRequest, sessId)
	g.JSON(http.StatusOK, gin.H{"op": BIND, "sessionId": sessId, "strict": k.webServer.UUID()})
}

type attachPodRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Container string `json:"container"`
	Shell     string `json:"shell"`
	ShellType string `json:"shellType"`
	Image     string `json:"image"`
	Cluster   string `json:"cluster"`
}

func (k *KesServer) Logs(
	cluster *k8s.Configure,
	namespace, name, container string,
	follow, previous, timestamps bool,
	sinceSeconds int64,
	sinceTime *time.Time,
	limitBytes int64,
	tailLines int64,
	out io.Writer,
) error {
	req := cluster.Clientset.
		CoreV1().
		RESTClient().
		Get().
		Namespace(namespace).
		Name(name).
		Resource("pods").
		SubResource("log").
		Param("container", container).
		Param("follow", strconv.FormatBool(follow)).
		Param("previous", strconv.FormatBool(previous)).
		Param("timestamps", strconv.FormatBool(timestamps))
	if sinceSeconds != 0 {
		req.Param("sinceSeconds", strconv.FormatInt(sinceSeconds, 10))
	}
	if sinceTime != nil {
		req.Param("sinceTime", sinceTime.Format(time.RFC3339))
	}
	if limitBytes != 0 {
		req.Param("limitBytes", strconv.FormatInt(limitBytes, 10))
	}
	if tailLines != 0 {
		req.Param("tailLines", strconv.FormatInt(tailLines, 10))
	}
	readerCloser, err := req.Stream(context.Background())
	if err != nil {
		return err
	}
	defer readerCloser.Close()
	_, err = io.Copy(out, readerCloser)

	return err

}

func SetUnstructuredListBaseCloudInfo(provider, region, az, cluster string, list *unstructured.UnstructuredList) {
	for _, item := range list.Items {
		annotations := item.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["github.com/ddx2x/provider"] = provider
		annotations["github.com/ddx2x/region"] = region
		annotations["github.com/ddx2x/az"] = az
		annotations["github.com/ddx2x/cluster"] = cluster
		item.SetAnnotations(annotations)
	}
}
