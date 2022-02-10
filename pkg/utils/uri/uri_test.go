package uri

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

type expectedStruct struct {
	method string
	uri    string
	Spec
}

func Test_parseOp(t *testing.T) {
	parser := NewURIParser()

	opMap := map[string]string{
		http.MethodGet: "view",
	}

	expectedStructs := []expectedStruct{
		{http.MethodGet, "/workload/api/v1/namespace/devops/pods/nginx-0-cg-0/container/nginx/op/attach", Spec{Service: "workload", Api: "api", Version: "v1", Resource: "pods", Namespace: "devops", Op: "attach", Name: "nginx-0-cg-0"}},
		{http.MethodGet, "/workload/api/v1/pods", Spec{Service: "workload", Api: "api", Namespace: "", Version: "v1", Resource: "pods", Name: "", Op: "view"}},
		{http.MethodGet, "/workload/api/v1/namespaces", Spec{Service: "workload", Api: "api", Namespace: "", Version: "v1", Resource: "namespaces", Name: "", Op: "view"}},
		{http.MethodGet, "/workload/api/v1/namespaces/im", Spec{Service: "workload", Api: "api", Namespace: "im", Version: "v1", Resource: "namespaces", Name: "im", Op: "view"}},
		{http.MethodGet, "/workload/api/v1/namespaces/im/op/label", Spec{Service: "workload", Api: "api", Namespace: "im", Version: "v1", Resource: "namespaces", Name: "im", Op: "label"}},
		{http.MethodGet, "/workload/api/v1/namespaces/im/pods", Spec{Service: "workload", Version: "v1", Api: "api", Namespace: "im", Resource: "pods", Name: "", Op: "view"}},
		{http.MethodGet, "/workload/api/v1/namespaces/im/pods/mypod", Spec{Service: "workload", Version: "v1", Api: "api", Namespace: "im", Resource: "pods", Name: "mypod", Op: "view"}},
		{http.MethodGet, "/workload/api/v1/namespaces/im/pods/mypod/op/log", Spec{Service: "workload", Version: "v1", Api: "api", Namespace: "im", Resource: "pods", Name: "mypod", Op: "log"}},
		{http.MethodGet, "/workload/apis/apps/v1/namespaces/ns1/deployments/op/watch?version=102222", Spec{Service: "workload", Namespace: "ns1", Version: "v1", Resource: "deployments", Op: "watch", Group: "apps", Api: "apis"}},
	}

	for _, value := range expectedStructs {
		if op, err := parser.ParseOp(value.method, value.uri, opMap); err != nil || !reflect.DeepEqual(&op.Spec, &value.Spec) {
			t.Fatalf("test parse uri (%s) error (%v) \nexpected (%s) \n   real （%s)", value.uri, err, &value.Spec, &op.Spec)
		}
	}

	_ = parser
	_ = expectedStructs
}

func Test_parseWatchPathAndResourceVersion(t *testing.T) {
	type expected struct {
		url         string
		expectedMap map[string]Content
	}
	expecteds := []expected{
		{url: "/watch?api=%2Fapis%2Fsystem.ddx2x.nip%2Fv1%2Fmenu%3Ftenant%3Dlizi%26watch%3D1%26resourceVersion%3D0",
			expectedMap: map[string]Content{
				"/apis/system.ddx2x.nip/v1/menu": Content{Cluster: "", ResourceVersion: "0", Tenant: "lizi"},
			},
		},
		{
			url: "/watch?api=%2Fapis%2Fsystem.ddx2x.nip%2Fv1%2Fcluster%3Fwatch%3D1%26resourceVersion%3D0&api=%2Fapis%2Fsystem.ddx2x.nip%2Fv1%2Fprovider%3Fwatch%3D1%26resourceVersion%3D0&api=%2Fapis%2Fsystem.ddx2x.nip%2Fv1%2Fregion%3Fwatch%3D1%26resourceVersion%3D0&api=%2Fapis%2Fsystem.ddx2x.nip%2Fv1%2Favailablezone%3Fwatch%3D1%26resourceVersion%3D0",
			expectedMap: map[string]Content{
				"/apis/system.ddx2x.nip/v1/cluster":       Content{Cluster: "", ResourceVersion: "0"},
				"/apis/system.ddx2x.nip/v1/provider":      Content{Cluster: "", ResourceVersion: "0"},
				"/apis/system.ddx2x.nip/v1/region":        Content{Cluster: "", ResourceVersion: "0"},
				"/apis/system.ddx2x.nip/v1/availablezone": Content{Cluster: "", ResourceVersion: "0"},
			},
		},
	}

	for _, expected := range expecteds {
		u, _ := url.Parse(expected.url)
		m, err := parseWatchPathAndResourceVersion(u)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected.expectedMap, m) {
			t.Fatalf("expected result not match url: %s map: %v", expected.url, expected.expectedMap)
		}
	}
}
