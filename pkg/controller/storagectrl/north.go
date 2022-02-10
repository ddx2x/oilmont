package storagectrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/compute"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (V *StorageCtrl) NorthOnAdd(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnAdd")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	storage := &compute.Storage{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, storage); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	unstructuredStorage, ok, err := V.checkObjStatusAndGetUnstructuredObj(storage, common.INIT)
	if !ok {
		if err != nil {
			flog.Warnf("unstructured storage %s error %v", storage.GetName(), err)
		}
		return
	}

	_, err = client.Interface.Resource(storageGvr).Namespace(unstructuredStorage.GetNamespace()).Create(
		context.Background(), unstructuredStorage, metav1.CreateOptions{})
	if err != nil {
		flog.Infof("create storage error %v", err)
		return
	}

	flog.Infof("create a new storage %s to k8s", unstructuredStorage.GetName())
	return
}

func (V *StorageCtrl) NorthOnUpdate(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnUpdate")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	storage := &compute.Storage{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, storage); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	unstructuredStorage, ok, err := V.checkObjStatusAndGetUnstructuredObj(storage, common.UPDATE)
	if !ok {
		if err != nil {
			flog.Warnf("unstructured storage %s error %v", storage.GetName(), err)
		}
		return
	}

	if _, _, err := client.Apply(context.Background(), storage.GetNamespace(), storageGvr, storage.GetName(), unstructuredStorage, false); err != nil {
		flog.Warnf("apply storage error %v", err)
		return
	}

	flog.Infof("update storage %s to k8s", unstructuredStorage.GetName())
	return
}

func (V *StorageCtrl) NorthOnDelete(obj core.IObject) {
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
	//_, _, err = client.Apply(context.Background(), storage.GetNamespace(), storageGvr, storage.GetName(),
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

func (V *StorageCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return V.stage.WatchEvent(ctx, common.DefaultDatabase, common.STORAGE, "0")
}

func (V *StorageCtrl) checkObjStatusAndGetUnstructuredObj(storage *compute.Storage, checkStatus string) (*unstructured.Unstructured, bool, error) {
	// 北桥接收到的数据，如果不符合要求状态，则不创建去k8s
	if storage.Spec.Status != checkStatus {
		return nil, false, nil
	}

	labels := make([]storageLabel, 0)
	for k, v := range storage.Labels {
		label := storageLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	storageParams := &StorageParams{
		Name:               storage.Name,
		Namespace:          storage.Namespace,
		Labels:             labels,
		Attachments:        storage.Spec.Attachments,
		CategoryType:       storage.Spec.CategoryType,
		DeleteWithInstance: storage.Spec.DeleteWithInstance,
		Description:        storage.Spec.Description,
		DiskChargeType:     storage.Spec.DiskChargeType,
		DiskType:           storage.Spec.DiskType,
		IOPS:               storage.Spec.IOPS,
		LocalName:          storage.Spec.LocalName,
		Message:            storage.Spec.Message,
		Region:             storage.Spec.Region,
		Zone:               storage.Spec.Zone,
		Size:               storage.Spec.Size,
		State:              storage.Spec.State,
		Status:             storage.Spec.Status,
		StorageId:          storage.Spec.StorageId,
		Throughput:         storage.Spec.Throughput,
	}

	unstructuredStorage, err := utilsObj.Render(storageParams, storageTpl)
	if err != nil {
		return nil, false, err
	}
	return unstructuredStorage, true, nil
}
