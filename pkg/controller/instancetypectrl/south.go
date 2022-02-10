package instancetypectrl

import (
	"context"
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func (V *InstanceTypeCtrl) SouthOnAdd(obj runtime.Object) {
	var name string
	flog := V.flog.WithField("func", "SouthOnAdd")

	instanceType := &unstructured.Unstructured{}
	err := utilsObj.Unmarshal(instanceType, obj)
	if err != nil {
		flog.Warnf("unmarshal instanceType data error: %v", obj)
		return
	}

	region := utilsObj.GetNestedString(instanceType.Object,
		"spec", "region")

	zone := utilsObj.GetNestedString(instanceType.Object,
		"spec", "zone")

	instanceTypeID := utilsObj.GetNestedString(instanceType.Object,
		"spec", "instanceTypeId")

	cores := utilsObj.GetNestedInt64(instanceType.Object,
		"spec", "cores")

	memory := utilsObj.GetNestedString(instanceType.Object,
		"spec", "memory")

	if instanceType.GetNamespace() == common.AWS {
		name = fmt.Sprintf("%s-%s", region, instanceTypeID)
	} else {
		name = fmt.Sprintf("%s-%s", zone, instanceTypeID)
	}

	instanceTypeObj := &system.InstanceType{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: instanceType.GetNamespace(),
			Workspace: common.DefaultWorkspace,
			Kind:      core.Kind(instanceType.GetKind()),
		},
		Spec: system.InstanceTypeSpec{
			Region: region,
			Cores:  cores,
			Memory: memory,
			ID:     instanceTypeID,
			Zone:   zone,
		},
	}

	err = V.stage.Get(common.DefaultDatabase, common.INSTANCETYPE, instanceTypeObj.GetName(), &system.InstanceType{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.INSTANCETYPE, instanceTypeObj)
		if createErr != nil {
			flog.Warnf("create instancetype error %v", createErr)
			return
		}
		flog.Infof("create instanceType %s, namespace %s", instanceTypeObj.GetName(), instanceTypeObj.GetNamespace())
		return
	}
	if err != nil {
		flog.Warnf("get instancetype error %v", err)
		return
	}
	_, update, err := V.stage.Apply(common.DefaultDatabase, common.INSTANCETYPE, instanceTypeObj.GetName(), instanceTypeObj, false)
	if err != nil {
		flog.Warnf("update instanceType error: %v", obj)
		return
	}
	if update {
		flog.Infof("update instanceType %s, namespace %s", instanceTypeObj.GetName(), instanceTypeObj.GetNamespace())
	}
}

func (V *InstanceTypeCtrl) SouthOnUpdate(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnUpdate")

	instanceType := &unstructured.Unstructured{}
	err := utilsObj.Unmarshal(instanceType, obj)
	if err != nil {
		flog.Warnf("unmarshal instanceType data error: %v", obj)
	}

	region := utilsObj.GetNestedString(instanceType.Object,
		"spec", "region")

	instanceTypeID := utilsObj.GetNestedString(instanceType.Object,
		"spec", "instanceTypeId")

	cores := utilsObj.GetNestedInt64(instanceType.Object,
		"spec", "cores")

	memory := utilsObj.GetNestedString(instanceType.Object,
		"spec", "memory")

	name := fmt.Sprintf("%s-%s", region, instanceTypeID)

	instanceTypeObj := &system.InstanceType{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: instanceType.GetNamespace(),
			Workspace: common.DefaultWorkspace,
			Kind:      core.Kind(instanceType.GetKind()),
		},
		Spec: system.InstanceTypeSpec{
			Region: region,
			Cores:  cores,
			Memory: memory,
			ID:     instanceTypeID,
		},
	}

	_, _, err = V.stage.Apply(common.DefaultDatabase, common.INSTANCETYPE, instanceTypeObj.GetName(), instanceTypeObj, false)

	if err != nil {
		flog.Warnf("update instanceType error: %v", obj)
	}
	flog.Infof("update new instanceType %s", instanceTypeObj.GetName())
}

func (V *InstanceTypeCtrl) SouthOnDelete(obj runtime.Object) {
	var name string

	flog := V.flog.WithField("func", "deleteImageToStage")
	instanceType := &unstructured.Unstructured{}
	err := utilsObj.Unmarshal(instanceType, obj)
	if err != nil {
		flog.Warnf("unmarshal instanceType data error: %v", obj)
	}

	region := utilsObj.GetNestedString(instanceType.Object,
		"spec", "region")
	zone := utilsObj.GetNestedString(instanceType.Object,
		"spec", "zone")
	instanceTypeID := utilsObj.GetNestedString(instanceType.Object,
		"spec", "instanceTypeId")

	if instanceType.GetNamespace() == common.AWS {
		name = fmt.Sprintf("%s-%s", region, instanceTypeID)
	} else {
		name = fmt.Sprintf("%s-%s", zone, instanceTypeID)
	}

	err = V.stage.Delete(common.DefaultDatabase, common.INSTANCETYPE, name, common.DefaultWorkspace)
	if err != nil {
		flog.Warnf("delete instanceType error: %v", obj)
	}
	flog.Infof("delete a instanceType %s from stage", instanceType.GetName())
}

func (V *InstanceTypeCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
	flog := V.flog.WithField("func", "SouthEventChs")
	flog.Info("instanceType ctrl start watch South Event")

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

		watchInterface, err := client.Interface.Resource(instanceTypeGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}
