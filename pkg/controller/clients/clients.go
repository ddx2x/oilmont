package clients

import (
	"context"
	"fmt"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	kubeVirtV1 "kubevirt.io/client-go/apis/core/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"
	"reflect"
	"sync"
	"time"
)

type KubeClient struct {
	ClientSet   *kubernetes.Clientset
	Interface   dynamic.Interface
	KubevirtCli kubecli.KubevirtClient
	Service     datasource.IDataSource
	Cfg         *rest.Config
}

type Clients struct {
	mutex       sync.Mutex
	kubeClients map[string]*KubeClient
}

func (kc *KubeClient) Apply(ctx context.Context, namespace string, gvr schema.GroupVersionResource, name string, unstructured *unstructured.Unstructured, forceUpdate bool) (newUnstructured *unstructured.Unstructured, isUpdate bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		getObj, getErr := kc.
			Interface.
			Resource(gvr).
			Namespace(namespace).
			Get(ctx, name, metav1.GetOptions{})

		if errors.IsNotFound(getErr) {
			newObj, createErr := kc.
				Interface.
				Resource(gvr).
				Namespace(namespace).
				Create(ctx, unstructured, metav1.CreateOptions{})
			newUnstructured = newObj
			return createErr
		}
		if getErr != nil {
			return getErr
		}

		kc.compareObject(getObj, unstructured, forceUpdate)

		newObj, updateErr := kc.
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

func (kc *KubeClient) ApplyDataVolume(ctx context.Context, namespace string, name string, volume *v1beta1.DataVolume, forceUpdate bool) (newObj *v1beta1.DataVolume, isUpdate bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		getObj, getErr := kc.
			KubevirtCli.
			CdiClient().CdiV1beta1().
			DataVolumes(namespace).
			Get(ctx, name, metav1.GetOptions{})

		if errors.IsNotFound(getErr) {
			_, createErr := kc.KubevirtCli.
				CdiClient().CdiV1beta1().
				DataVolumes(namespace).
				Create(ctx, volume, metav1.CreateOptions{})
			return createErr
		}
		if getErr != nil {
			return getErr
		}

		kc.compareDataVolume(volume, getObj, forceUpdate)
		var updateErr error
		newObj, updateErr = kc.
			KubevirtCli.
			CdiClient().CdiV1beta1().
			DataVolumes(namespace).
			Update(ctx, volume, metav1.UpdateOptions{})

		isUpdate = true
		return updateErr
	})

	return
}

func (kc *KubeClient) ApplyVMI( vmi *kubeVirtV1.VirtualMachineInstance, forceUpdate bool) (newObj *kubeVirtV1.VirtualMachineInstance, isUpdate bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		getObj, getErr := kc.
			KubevirtCli.
			VirtualMachineInstance(vmi.GetNamespace()).
			Get(vmi.GetName(), &metav1.GetOptions{})

		if errors.IsNotFound(getErr) {
			_, createErr := kc.KubevirtCli.
				VirtualMachineInstance(vmi.GetNamespace()).
				Create(vmi)
			return createErr
		}
		if getErr != nil {
			return getErr
		}

		kc.compareVMI(vmi, getObj, forceUpdate)
		var updateErr error
		newObj, updateErr = kc.
			KubevirtCli.
			VirtualMachineInstance(vmi.GetNamespace()).
			Update(vmi)

		isUpdate = true
		return updateErr
	})

	return
}

func (kc *KubeClient) compareDataVolume(obj *v1beta1.DataVolume, getObj *v1beta1.DataVolume, forceUpdate bool) {
	if !reflect.DeepEqual(obj.TypeMeta, getObj.TypeMeta) {
		obj.TypeMeta = getObj.TypeMeta
	}
	if !reflect.DeepEqual(obj.ObjectMeta, getObj.ObjectMeta) {
		obj.ObjectMeta = getObj.ObjectMeta
	}

	if forceUpdate {
		obj.Annotations["forceUpdate"] = fmt.Sprintf("%d", time.Now().Unix())
	}
}

func (kc *KubeClient) compareVMI(obj *kubeVirtV1.VirtualMachineInstance, getObj *kubeVirtV1.VirtualMachineInstance, forceUpdate bool) {
	if !reflect.DeepEqual(obj.TypeMeta, getObj.TypeMeta) {
		obj.TypeMeta = getObj.TypeMeta
	}
	if !reflect.DeepEqual(obj.ObjectMeta, getObj.ObjectMeta) {
		obj.ObjectMeta = getObj.ObjectMeta
	}

	if forceUpdate {
		obj.Annotations["forceUpdate"] = fmt.Sprintf("%d", time.Now().Unix())
	}
}

func (kc *KubeClient) compareObject(getObj, obj *unstructured.Unstructured, forceUpdate bool) {
	if !reflect.DeepEqual(getObj.Object["metadata"], obj.Object["metadata"]) {
		getObj.Object["metadata"] = kc.compareMetadataLabelsOrAnnotation(
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

func (kc *KubeClient) compareMetadataLabelsOrAnnotation(old, new map[string]interface{}) map[string]interface{} {
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

func NewClients() *Clients {
	return &Clients{
		mutex:       sync.Mutex{},
		kubeClients: make(map[string]*KubeClient),
	}
}

func (cs *Clients) AddKubeClient(name string, c *KubeClient) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.kubeClients[name] = c
}

func (cs *Clients) RemoveClient(c string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	delete(cs.kubeClients, c)
}

func (cs *Clients) GetClient(name string) (*KubeClient, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	kubeClient, exist := cs.kubeClients[name]
	if !exist {
		return nil, fmt.Errorf("not found config client %s", name)
	}
	return kubeClient, nil
}
