package kes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/ddx2x/oilmont/pkg/api"
	"github.com/ddx2x/oilmont/pkg/common"
)

const PrometheusAddress = "http://prometheus.kube-system.svc.cluster.local/api/v1/query_range"

func (k *KesServer) Proxy(ctx context.Context, cluster string, params map[string]string, body []byte) (map[string]interface{}, error) {
	var bodyMap map[string]string
	var resultMap = make(map[string]interface{})
	err := json.Unmarshal(body, &bodyMap)
	if err != nil {
		return nil, err
	}
	if cluster == "" {
		cluster = "default"
	}

	cli := k.multiCluster.Get(cluster)
	if cli == nil {
		return nil, fmt.Errorf("request cluster %s is not found", cluster)
	}

	if common.InCluster {
		for bodyKey, bodyValue := range bodyMap {
			resp, err := resty.New().R().
				SetQueryParams(params).
				SetQueryParam("query", bodyValue).
				Get(PrometheusAddress)
			if err != nil {
				return nil, err
			}
			var KesServerContextMap map[string]interface{}
			err = json.Unmarshal([]byte(resp.String()), &KesServerContextMap)
			if err != nil {
				return nil, err
			}
			resultMap[bodyKey] = KesServerContextMap
		}
		return resultMap, nil
	}

	for bodyKey, bodyValue := range bodyMap {
		req := cli.Clientset.RESTClient().
			Get().
			Namespace("kube-system").
			Resource("services").
			Name("prometheus:80").
			SubResource("proxy").
			Suffix("api/v1/query_range")

		for k, v := range params {
			req.Param(k, v)
		}

		req.Param("query", bodyValue)

		raw, err := req.DoRaw(ctx)
		if err != nil {
			return nil, err
		}

		var KesServerContextMap map[string]interface{}
		err = json.Unmarshal(raw, &KesServerContextMap)
		if err != nil {
			return nil, err
		}
		resultMap[bodyKey] = KesServerContextMap
	}

	return resultMap, nil
}

func (k *KesServer) PodMetrics(ctx context.Context, cluster, namespace, name string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	uri := fmt.Sprintf("apis/metrics.k8s.io/v1beta1/%s/%s/pods", namespace, name)
	if cluster == "" {
		cluster = "default"
	}

	cli := k.multiCluster.Get(cluster)
	if cli == nil {
		return nil, fmt.Errorf("request cluster %s is not found", cluster)
	}

	data, err := cli.RESTClient().Get().AbsPath(uri).DoRaw(ctx)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (k *KesServer) NamespacePodMetrics(ctx context.Context, cluster, namespace string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	uri := "apis/metrics.k8s.io/v1beta1/pods"
	if namespace != "" {
		uri = fmt.Sprintf("apis/metrics.k8s.io/v1beta1/namespaces/%s/pods", namespace)
	}

	if cluster == "" {
		cluster = "default"
	}

	cli := k.multiCluster.Get(cluster)
	if cli == nil {
		return nil, fmt.Errorf("request cluster %s is not found", cluster)
	}

	data, err := cli.RESTClient().Get().AbsPath(uri).DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (k *KesServer) Metrics(g *gin.Context) {
	body, err := g.GetRawData()
	if err != nil {
		api.RequestParametersError(g, fmt.Errorf("params not obtain or params parse error: %s", err))
		return
	}
	params := make(map[string]string)
	params["start"] = g.Query("start")
	params["end"] = g.Query("end")
	params["step"] = g.Query("step")
	params["kubernetes_namespace"] = g.Query("kubernetes_namespace")
	params["cluster"] = g.Query("cluster")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	bufRaw, err := k.Proxy(ctx, g.Query("cluster"), params, body)
	if err != nil {
		api.InternalServerError(g, "", err)
		return
	}

	g.JSON(http.StatusOK, bufRaw)
}

func (k *KesServer) NodeMetrics(ctx context.Context, cluster string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if cluster == "" {
		cluster = "default"
	}
	cli := k.multiCluster.Get(cluster)
	if cli == nil {
		return nil, fmt.Errorf("request cluster %s is not found", cluster)
	}

	data, err := cli.RESTClient().Get().
		AbsPath("apis/metrics.k8s.io/v1beta1/nodes").
		DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
