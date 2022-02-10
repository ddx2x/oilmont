package k8s

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Resource struct {
	Name   string
	Schema schema.GroupVersionResource
}

type ResourceRegistry interface {
	Register(res string, gvr schema.GroupVersionResource)
	Subscript(d dynamicinformer.DynamicSharedInformerFactory, stop <-chan struct{})
	GetGVR(res string) (schema.GroupVersionResource, error)
	Length() int
}

type ILister interface {
	List(ctx context.Context, namespace string, resourceType string, selector string) (*unstructured.UnstructuredList, error)
	Get(ctx context.Context, namespace string, resourceType string, name string, subresources ...string) (*unstructured.Unstructured, error)
	ListWithGVR(ctx context.Context, namespace string, gvr schema.GroupVersionResource, selector string) (*unstructured.UnstructuredList, error)
	ListWithPage(ctx context.Context, namespace string, resourceType string, flag string, pos, size int64, selector string) (*unstructured.UnstructuredList, error)
}

type IWatcher interface {
	Watch(ctx context.Context, namespace string, resourceType string, resourceVersion string, selector string) (eventCh <-chan watch.Event, stop func(), err error)
}

type IUpdater interface {
	ApplyWithGVR(ctx context.Context, namespace, name string, gvr *schema.GroupVersionResource, unstructured *unstructured.Unstructured) (newUnstructured *unstructured.Unstructured, isUpdate bool, err error)
	Apply(ctx context.Context, namespace string, resourceType string, name string, unstructured *unstructured.Unstructured, forceUpdate bool) (newUnstructured *unstructured.Unstructured, isUpdate bool, err error)
	Delete(ctx context.Context, namespace string, resourceType string, name string) error
	Patch(ctx context.Context, namespace string, resourceType string, name string, data []byte) (*unstructured.Unstructured, error)
}

type IRESTClient interface {
	RESTClient() rest.Interface
	ClientSet() *kubernetes.Clientset
}

type IDiscoveryClient interface {
	DiscoveryClient() *discovery.DiscoveryClient
}

type Interface interface {
	ILister
	IWatcher
	IUpdater
	IRESTClient
	IDiscoveryClient
}
