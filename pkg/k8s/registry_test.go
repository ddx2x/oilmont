package k8s

import (
	"testing"
)

func Test_In(t *testing.T) {
	res := Resources{
		{"pods", GVR{Group: "", Version: "v1", Resource: "pods"}},
		{"namespaces", GVR{Group: "", Version: "v1", Resource: "namespaces"}},
	}

	if !res.In("pods") {
		t.Failed()
	}

	if res.In("stones") {
		t.Failed()
	}

}
