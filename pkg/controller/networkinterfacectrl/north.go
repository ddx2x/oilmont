package networkinterfacectrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/networking"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (V *NetworkInterfaceCtrl) NorthOnAdd(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnAdd")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	networkInterface := &networking.NetworkInterface{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, networkInterface); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	unstructuredENI, ok, err := V.checkObjStatusAndGetUnstructuredObj(networkInterface, common.INIT)
	if !ok {
		if err != nil {
			flog.Warnf("unstructured networkInterface %s error %v", networkInterface.GetName(), err)
		}
		return
	}

	_, err = client.Interface.Resource(NetworkInterfaceGvr).Namespace(unstructuredENI.GetNamespace()).Create(
		context.Background(), unstructuredENI, metav1.CreateOptions{})
	if err != nil {
		flog.Infof("create networkInterface error %v", err)
		return
	}

	flog.Infof("create a new networkInterface %s to k8s", unstructuredENI.GetName())
	return
}

func (V *NetworkInterfaceCtrl) NorthOnUpdate(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnUpdate")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	networkInterface := &networking.NetworkInterface{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, networkInterface); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	unstructuredENI, ok, err := V.checkObjStatusAndGetUnstructuredObj(networkInterface, common.UPDATE)
	if !ok {
		if err != nil {
			flog.Warnf("unstructured networkInterface %s error %v", networkInterface.GetName(), err)
		}
		return
	}

	_, err = client.Interface.Resource(NetworkInterfaceGvr).Namespace(unstructuredENI.GetNamespace()).Update(
		context.Background(), unstructuredENI, metav1.UpdateOptions{})

	if err != nil {
		flog.Infof("update networkInterface obj error %v", err)
		return
	}

	flog.Infof("update networkInterface %s to k8s", unstructuredENI.GetName())
	return
}

func (V *NetworkInterfaceCtrl) NorthOnDelete(obj core.IObject) {
	//flog := V.flog.WithField("func", "NorthOnDelete")
	//
	//client, err := V.cs.GetClient(common.DefaultKubernetes)
	//if err != nil {
	//	flog.Infof("get client error %v", err)
	//	return
	//}
	//
	//storage := &compute.Storage{}
	//if err := utilsObj.UnstructuredObjectToInstanceObj(obj, storage); err != nil {
	//	flog.Warnf("unstructured obj error %v", err)
	//	return
	//}
	//
	//unstructuredStorage, ok, err := V.checkObjStatusAndGetUnstructuredObj(storage, storage.Spec.Status) // 删除时不需要关注 status
	//if !ok {
	//	if err != nil {
	//		flog.Warnf("unstructured storage %s error %v", storage.GetName(), err)
	//	}
	//	return
	//}
	//
	//_, _, err = client.Apply(context.Background(), storage.GetNamespace(), NetworkInterfaceGvr, storage.GetName(),
	//	unstructuredStorage, false)
	//
	//if err != nil {
	//	flog.Infof("delete storage error %v", err)
	//	return
	//}
	//
	//flog.Infof("trying delete storage %s from k8s", storage.GetName())
	return
}

func (V *NetworkInterfaceCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return V.stage.WatchEvent(ctx, common.DefaultDatabase, common.NETWORKINTERFACE, "0")
}

func (V *NetworkInterfaceCtrl) checkObjStatusAndGetUnstructuredObj(networkInterface *networking.NetworkInterface, checkStatus string) (*unstructured.Unstructured, bool, error) {
	// 北桥接收到的数据，如果不符合要求状态，则不创建去k8s
	if networkInterface.Spec.Status != checkStatus {
		return nil, false, nil
	}

	labels := make([]eniLabel, 0)
	for k, v := range networkInterface.Labels {
		label := eniLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	networkInterfaceParams := &NetworkInterfaceParams{
		Name:      networkInterface.Name,
		Namespace: networkInterface.Namespace,
		Labels:    labels,
	}

	unstructuredStorage, err := utilsObj.Render(networkInterfaceParams, networkInterfaceTpl)
	if err != nil {
		return nil, false, err
	}
	return unstructuredStorage, true, nil
}
