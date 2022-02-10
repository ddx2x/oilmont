package k8s

import (
	"context"
	"fmt"

	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

var _ Interface = &dataSource{}

func NewInterface(configure *Configure) Interface {
	return &dataSource{configure}
}

type dataSource struct {
	*Configure
}

func (d *dataSource) ApplyWithGVR(ctx context.Context, namespace, name string, gvr *schema.GroupVersionResource, unstructured *unstructured.Unstructured) (newUnstructured *unstructured.Unstructured, isUpdate bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		getObj, getErr := d.
			Interface.
			Resource(*gvr).
			Namespace(namespace).
			Get(ctx, name, metav1.GetOptions{})

		if errors.IsNotFound(getErr) {
			newObj, createErr := d.
				Interface.
				Resource(*gvr).
				Namespace(namespace).
				Create(ctx, unstructured, metav1.CreateOptions{})
			newUnstructured = newObj
			return createErr
		}
		if getErr != nil {
			return getErr
		}

		d.compareObject(getObj, unstructured, false)

		newObj, updateErr := d.
			Interface.
			Resource(*gvr).
			Namespace(namespace).
			Update(ctx, getObj, metav1.UpdateOptions{})

		newUnstructured = newObj
		isUpdate = true
		return updateErr
	})

	return
}

func (d *dataSource) DiscoveryClient() *discovery.DiscoveryClient {
	return d.Configure.DiscoveryClient
}

func (d *dataSource) ListWithGVR(ctx context.Context, namespace string, gvr schema.GroupVersionResource, selector string) (*unstructured.UnstructuredList, error) {
	return d.Interface.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
}

func (d *dataSource) RESTClient() rest.Interface {
	return d.Configure.RESTClient()
}

func (d *dataSource) ClientSet() *kubernetes.Clientset {
	return d.Configure.Clientset
}

func (d *dataSource) XGet(namespace string, resourceType string, name string) (*unstructured.Unstructured, error) {
	gvr, err := d.GetGVR(resourceType)
	if err != nil {
		return nil, err
	}
	obj, err := d.ForResource(gvr).Lister().ByNamespace(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: unstructuredObj}, nil
}

func (d *dataSource) Watch(ctx context.Context, namespace string, resourceType string, resourceVersion string, selector string) (<-chan watch.Event, func(), error) {
	gvr, err := d.GetGVR(resourceType)
	if err != nil {
		return nil, nil, err
	}
	timeoutSeconds := int64(0)
	listOptions := metav1.ListOptions{
		ResourceVersion: resourceVersion,
		LabelSelector:   selector,
		TimeoutSeconds:  &timeoutSeconds,
	}
	watchInterface, err := d.Interface.Resource(gvr).Namespace(namespace).Watch(ctx, listOptions)
	if err != nil {
		return nil, nil, err
	}
	return watchInterface.ResultChan(), watchInterface.Stop, nil
}

func (d *dataSource) Apply(ctx context.Context, namespace string, resourceType string, name string, unstructured *unstructured.Unstructured, forceUpdate bool) (newUnstructured *unstructured.Unstructured, isUpdate bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		gvr, err := d.GetGVR(resourceType)
		if err != nil {
			return err
		}
		getObj, getErr := d.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})

		if errors.IsNotFound(getErr) {
			newObj, createErr := d.Resource(gvr).Namespace(namespace).Create(ctx, unstructured, metav1.CreateOptions{})
			newUnstructured = newObj
			return createErr
		}
		if getErr != nil {
			return getErr
		}

		d.compareObject(getObj, unstructured, forceUpdate)

		newObj, updateErr := d.
			Interface.
			Resource(gvr).
			Namespace(namespace).
			Update(ctx, getObj, metav1.UpdateOptions{})

		newUnstructured = newObj
		isUpdate = true
		return updateErr
	})

	return
}

func (d *dataSource) Delete(ctx context.Context, namespace string, resourceType string, name string) error {
	gvr, err := d.GetGVR(resourceType)
	if err != nil {
		return err
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return d.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	})
}

func (d *dataSource) Patch(ctx context.Context, namespace string, resourceType string, name string, data []byte) (result *unstructured.Unstructured, err error) {
	gvr, err := d.GetGVR(resourceType)
	if err != nil {
		return nil, err
	}
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		result, err = d.Resource(gvr).
			Namespace(namespace).
			Patch(ctx, name, apimachinerytypes.MergePatchType, data, metav1.PatchOptions{})
		return err
	})
	return
}

func (d *dataSource) ListWithPage(ctx context.Context, namespace string, resourceType string, flag string, pos, size int64, selector string) (*unstructured.UnstructuredList, error) {
	gvr, err := d.ResourceRegistry.GetGVR(resourceType)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{LabelSelector: selector}
	if flag != "" {
		opts.Continue = flag
	}
	if size > 0 {
		opts.Limit = size + pos
	}
	return d.Resource(gvr).Namespace(namespace).List(ctx, opts)
}

func (d *dataSource) List(ctx context.Context, namespace string, resourceType string, selector string) (*unstructured.UnstructuredList, error) {
	gvr, err := d.GetGVR(resourceType)
	if err != nil {
		return nil, err
	}
	return d.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
}

func (d *dataSource) Get(ctx context.Context, namespace string, resourceType string, name string, subresources ...string) (*unstructured.Unstructured, error) {
	gvr, err := d.GetGVR(resourceType)
	if err != nil {
		return nil, err
	}
	return d.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{}, subresources...)
}

func (d *dataSource) compareObject(getObj, obj *unstructured.Unstructured, forceUpdate bool) {
	if !reflect.DeepEqual(getObj.Object["metadata"], obj.Object["metadata"]) {
		getObj.Object["metadata"] = d.compareMetadataLabelsORAnnotation(
			getObj.Object["metadata"].(map[string]interface{}),
			obj.Object["metadata"].(map[string]interface{}),
		)
	}

	if forceUpdate {
		metadata := getObj.Object["metadata"].(map[string]interface{})
		if metadata == nil {
			goto NEXT0
		}

		annotations, exist := metadata["annotations"]
		if !exist {
			annotations = make(map[string]interface{})
		}

		annotationsMap := annotations.(map[string]interface{})
		annotationsMap["forceUpdate"] = fmt.Sprintf("%d", time.Now().Unix())
		metadata["annotations"] = annotationsMap
		getObj.Object["metadata"] = metadata
	}

NEXT0:
	if !reflect.DeepEqual(getObj.Object["spec"], obj.Object["spec"]) {
		getObj.Object["spec"] = obj.Object["spec"]
	}

	// configMap
	if !reflect.DeepEqual(getObj.Object["data"], obj.Object["data"]) {
		getObj.Object["data"] = obj.Object["data"]
	}

	if !reflect.DeepEqual(getObj.Object["binaryData"], obj.Object["binaryData"]) {
		getObj.Object["binaryData"] = obj.Object["binaryData"]
	}

	if !reflect.DeepEqual(getObj.Object["stringData"], obj.Object["stringData"]) {
		getObj.Object["stringData"] = obj.Object["stringData"]
	}

	if !reflect.DeepEqual(getObj.Object["type"], obj.Object["type"]) {
		getObj.Object["type"] = obj.Object["type"]
	}

	if !reflect.DeepEqual(getObj.Object["secrets"], obj.Object["secrets"]) {
		getObj.Object["ddx2xs"] = obj.Object["secrets"]
	}

	if !reflect.DeepEqual(getObj.Object["imagePullSecrets"], obj.Object["imagePullSecrets"]) {
		getObj.Object["imagePullSecrets"] = obj.Object["imagePullSecrets"]
	}
	// storageClass field
	if !reflect.DeepEqual(getObj.Object["provisioner"], obj.Object["provisioner"]) {
		getObj.Object["provisioner"] = obj.Object["provisioner"]
	}

	if !reflect.DeepEqual(getObj.Object["parameters"], obj.Object["parameters"]) {
		getObj.Object["parameters"] = obj.Object["parameters"]
	}

	if !reflect.DeepEqual(getObj.Object["reclaimPolicy"], obj.Object["reclaimPolicy"]) {
		getObj.Object["reclaimPolicy"] = obj.Object["reclaimPolicy"]
	}

	if !reflect.DeepEqual(getObj.Object["volumeBindingMode"], obj.Object["volumeBindingMode"]) {
		getObj.Object["volumeBindingMode"] = obj.Object["volumeBindingMode"]
	}
}

func (d *dataSource) compareMetadataLabelsORAnnotation(old, new map[string]interface{}) map[string]interface{} {
	newLabels, exist := new["labels"]
	if exist {
		old["labels"] = newLabels
	}
	newAnnotations, exist := new["annotations"]
	if exist {
		old["annotations"] = newAnnotations
	}

	newOwnerReferences, exist := new["ownerReferences"]
	if exist {
		old["ownerReferences"] = newOwnerReferences
	}
	return old
}
