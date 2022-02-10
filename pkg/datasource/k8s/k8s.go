package k8s

import (
	"context"
	"fmt"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"reflect"
	"time"
)

var _ datasource.IDataSource = &IDataSourceImpl{}

type IDataSourceImpl struct {
	Interface dynamic.Interface
}

func NewIDataSource(Interface dynamic.Interface) *IDataSourceImpl {
	return &IDataSourceImpl{Interface: Interface}
}

func (i *IDataSourceImpl) List(namespace string, gvr schema.GroupVersionResource, flag string, pos, size int64, selector interface{}) (*unstructured.UnstructuredList, error) {
	var err error
	var items *unstructured.UnstructuredList
	opts := metav1.ListOptions{}

	if selector == nil || selector == "" {
		selector = labels.Everything()
	}
	switch selector.(type) {
	case labels.Selector:
		opts.LabelSelector = selector.(labels.Selector).String()
	case string:
		if selector != "" {
			opts.LabelSelector = selector.(string)
		}
	}

	if flag != "" {
		opts.Continue = flag
	}
	if size > 0 {
		opts.Limit = size + pos
	}

	items, err = i.
		Interface.
		Resource(gvr).
		Namespace(namespace).
		List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (i *IDataSourceImpl) Get(namespace, name string, gvr schema.GroupVersionResource, subresources ...string) (*unstructured.Unstructured, error) {
	object, err := i.
		Interface.
		Resource(gvr).
		Namespace(namespace).
		Get(context.Background(), name, metav1.GetOptions{}, subresources...)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (i *IDataSourceImpl) Apply(namespace, name string, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, forceUpdate bool) (result *unstructured.Unstructured, isUpdate bool, err error) {
	retryErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {

		ctx := context.Background()
		getObj, getErr := i.
			Interface.
			Resource(gvr).
			Namespace(namespace).
			Get(ctx, name, metav1.GetOptions{})

		if errors.IsNotFound(getErr) {
			newObj, createErr := i.
				Interface.
				Resource(gvr).
				Namespace(namespace).
				Create(ctx, obj, metav1.CreateOptions{})
			result = newObj
			return createErr
		}

		if getErr != nil {
			return getErr
		}

		compareObject(getObj, obj, forceUpdate)

		newObj, updateErr := i.
			Interface.
			Resource(gvr).
			Namespace(namespace).
			Update(ctx, getObj, metav1.UpdateOptions{})

		result = newObj
		isUpdate = true
		return updateErr
	})
	err = retryErr

	return
}

func (i *IDataSourceImpl) Delete(namespace, name string, gvr schema.GroupVersionResource) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return i.
			Interface.
			Resource(gvr).
			Namespace(namespace).
			Delete(context.Background(), name, metav1.DeleteOptions{})
	})
}

func (i *IDataSourceImpl) Watch(namespace string, resource, resourceVersion string, timeoutSeconds int64, selector interface{}) (<-chan watch.Event, error) {
	panic("implement me")
}

func compareObject(getObj, obj *unstructured.Unstructured, forceUpdate bool) {
	if !reflect.DeepEqual(getObj.Object["metadata"], obj.Object["metadata"]) {
		getObj.Object["metadata"] = compareMetadataLabelsOrAnnotation(
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
		getObj.Object["secrets"] = obj.Object["secrets"]
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

func compareMetadataLabelsOrAnnotation(old, new map[string]interface{}) map[string]interface{} {
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
