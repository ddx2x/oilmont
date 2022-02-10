package storagectrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/compute"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	objUtils "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func (V *StorageCtrl) SouthOnAdd(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnAdd")
	unstructuredStorage := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(unstructuredStorage, obj)
	if err != nil {
		flog.Warnf("unmarshal storage data error: %v", obj)
	}

	storage, ok := V.checkObjStatusAndGetObj(unstructuredStorage, common.INIT)
	if !ok {
		return
	}

	err = V.stage.Get(common.DefaultDatabase, common.STORAGE, storage.GetName(), &compute.Storage{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.STORAGE, storage)
		if createErr != nil {
			flog.Warnf("create storage error: %v", createErr)
			return
		}
		flog.Infof("create storage %s,namespace: %s, workspace: %s", storage.GetName(), storage.GetNamespace(), storage.GetWorkspace())
		return
	}
	if err != nil {
		flog.Warnf("get storage error: %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.STORAGE, storage.GetName(), storage, false)
	if err != nil {
		flog.Warnf("create storage error: %v", err)
		return
	}
	if update {
		flog.Infof("update storage %s, workspace: %s, namespace: %s", storage.GetName(), storage.GetWorkspace(), storage.GetNamespace())
	}
	return
}

func (V *StorageCtrl) SouthOnUpdate(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnUpdate")
	var force = false
	unstructuredStorage := &unstructured.Unstructured{}
	err := objUtils.Unmarshal(unstructuredStorage, obj)
	if err != nil {
		flog.Warnf("unmarshal storage data error: %v", obj)
	}

	storage, ok := V.checkObjStatusAndGetObj(unstructuredStorage, common.DELETE, common.UPDATE)
	if !ok {
		return
	}

	if storage.Spec.Status == common.FAIL {
		force = true
		storage.Metadata.IsDelete = false
	}

	_, _, err = V.stage.Apply(common.DefaultDatabase, common.STORAGE, storage.GetName(), storage, force)
	if err != nil {
		flog.Warnf("update storage error: %v", obj)
		return
	}

	flog.Infof("update a storage %s to stage", storage.GetName())
	return
}

func (V *StorageCtrl) SouthOnDelete(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnDelete")
	unstructuredStorage := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(unstructuredStorage, obj)
	if err != nil {
		flog.Warnf("unmarshal storage data error: %v", obj)
	}

	labels := unstructuredStorage.GetLabels()

	err = V.stage.Delete(common.DefaultDatabase, common.STORAGE, unstructuredStorage.GetName(), labels["workspace"])
	if err != nil {
		flog.Warnf("delete storage error: %v", obj)
		return
	}

	flog.Warnf("delete a storage %s from stage", unstructuredStorage.GetName())
	return
}

func (V *StorageCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
	clusters := make([]system.Cluster, 0)
	if err := V.stage.ListToObject(common.DefaultDatabase, common.CLUSTER, map[string]interface{}{}, &clusters, true); err != nil {
		return nil, err
	}

	channels := make([]<-chan watch.Event, 0)
	for _, cluster := range clusters {
		client, err := V.cs.GetClient(cluster.GetName())
		if err != nil {
			return nil, err
		}

		watchInterface, err := client.Interface.Resource(storageGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}

func (V *StorageCtrl) checkObjStatusAndGetObj(unstructuredStorage *unstructured.Unstructured, checkStatus ...string) (*compute.Storage, bool) {
	status := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "status")
	for _, s := range checkStatus {
		if status == s {
			return nil, false
		}
	}

	categoryType := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "categoryType")
	deleteWithInstance := objUtils.GetNestedBool(unstructuredStorage.Object,
		"spec", "deleteWithInstance")
	description := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "description")
	diskChargeType := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "diskChargeType")
	diskType := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "diskType")
	iops := objUtils.GetNestedInt(unstructuredStorage.Object,
		"spec", "iops")
	localName := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "localName")
	message := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "message")
	region := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "region")
	zone := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "zone")
	size := objUtils.GetNestedInt(unstructuredStorage.Object,
		"spec", "size")
	throughput := objUtils.GetNestedInt(unstructuredStorage.Object,
		"spec", "throughput")
	state := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "state")
	storageId := objUtils.GetNestedString(unstructuredStorage.Object,
		"spec", "storageId")

	spec := objUtils.GetNestedMap(unstructuredStorage.Object,
		"spec")

	labels := map[string]interface{}{}
	getLabels := unstructuredStorage.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	storage := &compute.Storage{
		Metadata: core.Metadata{
			Name:      unstructuredStorage.GetName(),
			Namespace: unstructuredStorage.GetNamespace(),
			Workspace: getLabels["workspace"],
			Kind:      core.Kind(unstructuredStorage.GetKind()),
			Labels:    labels,
		}, Spec: compute.StorageSpec{
			CategoryType:       categoryType,
			DeleteWithInstance: deleteWithInstance,
			Description:        description,
			DiskChargeType:     diskChargeType,
			IOPS:               iops,
			LocalName:          localName,
			Message:            message,
			Region:             region,
			DiskType:           diskType,
			Zone:               zone,
			Size:               size,
			State:              state,
			Throughput:         throughput,
			Status:             status,
			StorageId:          storageId,
		},
	}
	attachments := spec["attachments"]
	if attachments != nil {
		for _, obj := range attachments.([]interface{}) {
			_attachment := obj.(map[string]interface{})
			attachment := compute.Attachment{
				AttachedTime: _attachment["attachedTime"].(string),
				Device:       _attachment["device"].(string),
				InstanceId:   _attachment["instanceId"].(string),
				State:        _attachment["state"].(string),
			}
			storage.Spec.Attachments = append(storage.Spec.Attachments, attachment)
		}
	}

	if storage.Workspace == "" {
		storage.Workspace = common.DefaultWorkspace
	}
	return storage, true
}
