package dict

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func Test_shift(t *testing.T) {
	path := "a.b.c"
	prefix, remain := shift(path)
	if prefix != "a" || remain != "b.c" {
		t.Fatal("expected not equal")
	}
}

func Test_Delete(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": 123,
			},
		},
	}
	expected := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{},
		},
	}
	Delete(data, "a.b.c")

	if !reflect.DeepEqual(data, expected) {
		t.Fatal("expected not equal")
	}
}

func Test_Set(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": 123,
			},
		},
	}
	expected := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": 456,
			},
		},
	}

	Set(data, "a.b.c", 456)

	if !reflect.DeepEqual(data, expected) {
		t.Fatal("expected not equal")
	}
}

func Test_Get(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": 123,
			},
		},
	}
	value := Get(data, "a.b.c")
	if value.(int) != 123 {
		t.Fatal("expected not equal")
	}
}

const deployment = `
kind: Deployment
apiVersion: apps/v1
metadata:
  name: mall-riskcontrol-jobs-c15f4020-tm
  namespace: flink-operator
  selfLink: >-
    /apis/apps/v1/namespaces/flink-operator/deployments/mall-riskcontrol-jobs-c15f4020-tm
  uid: 0ca70827-2134-4caa-a618-e90da3de495c
  resourceVersion: '214651587'
  generation: 1
  creationTimestamp: '2021-02-02T09:47:10Z'
  labels:
    environment: development
    flink-app: mall-riskcontrol-jobs
    flink-app-hash: c15f4020
    flink-deployment-type: taskmanager
  annotations:
    deployment.kubernetes.io/revision: '1'
    flink-job-properties: |-
      jarName: mall-riskcontrol-jobs-1.0.0-SNAPSHOT.jar
      parallelism: 5
      entryClass:cn.ecpark.service.riskcontrol.mall.jobs.order.OrderAvgJob
      programArgs:""
    kubectl.kubernetes.io/last-applied-configuration: >
      {"apiVersion":"flink.k8s.io/v1beta1","kind":"FlinkApplication","metadata":{"annotations":{},"labels":{"environment":"development"},"name":"mall-riskcontrol-jobs","namespace":"flink-operator"},"spec":{"deleteMode":"None","entryClass":"cn.ecpark.service.riskcontrol.mall.jobs.order.OrderAvgJob","flinkConfig":{"state.backend.fs.checkpointdir":"file:///checkpoints/flink/checkpoints","state.checkpoints.dir":"file:///checkpoints/flink/externalized-checkpoints","state.savepoints.dir":"file:///checkpoints/flink/savepoints","web.upload.dir":"/opt/flink"},"flinkVersion":"1.8","image":"harbor.ym/devops/mall-riskcontrol-jobs:v0.0.1","jarName":"mall-riskcontrol-jobs-1.0.0-SNAPSHOT.jar","jobManagerConfig":{"envConfig":{"env":[{"name":"ENV","value":"DEV"}]},"replicas":1,"resource":{"requests":{"cpu":"0.1","memory":"2Gi"}}},"parallelism":5,"taskManagerConfig":{"resource":{"requests":{"cpu":"0.1","memory":"2Gi"}},"taskSlots":2}}}
  ownerReferences:
    - apiVersion: flink.k8s.io/v1beta1
      kind: FlinkApplication
      name: mall-riskcontrol-jobs
      uid: 0428aa4e-81b6-4f08-8625-6fb1c5a42bfe
      controller: true
      blockOwnerDeletion: true
spec:
  replicas: 3
  selector:
    matchLabels:
      environment: development
      flink-app: mall-riskcontrol-jobs
      flink-app-hash: c15f4020
      flink-deployment-type: taskmanage
`

func Test_Get_Deployment(t *testing.T) {
	object := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(deployment), &object)
	if err != nil {
		t.Fatal(err)
	}
	v := Get(object, "spec.selector.matchLabels")
	if v == nil {
		t.Fatal("get expected data is nil")
	}
	fmt.Printf("%s\n", v)

}

var src = map[string]interface{}{
	"metadata": map[string]string{"A": "B"},
	"spec": map[string]interface{}{
		"ops": []string{"c", "r", "u", "d"},
		"a":   1,
	},
}

var dest = map[string]interface{}{
	"metadata": map[string]string{"A": "B"},
	"spec": map[string]interface{}{
		"ops": []string{"c", "r", "u", "d"},
		"a":   2,
		"b":   2,
		"c":   3,
	},
}

func Test_Compare(t *testing.T) {
	CompareMergeObject(src, dest, "spec.a")
	expected := map[string]interface{}{
		"metadata": map[string]string{"A": "B"},
		"spec": map[string]interface{}{
			"ops": []string{"c", "r", "u", "d"},
			"a":   2,
		},
	}

	if !reflect.DeepEqual(src, expected) {
		t.Fatalf("expected equal,%v", src)
	}

	CompareMergeObject(src, dest, "spec")

	if !reflect.DeepEqual(src, dest) {
		t.Fatalf("expected equal,%v", src)
	}
}
