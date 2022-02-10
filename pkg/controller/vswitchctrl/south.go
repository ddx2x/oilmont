package vswitchctrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/networking"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	objUtils "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func (V *VSwitchCtrl) SouthOnAdd(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnAdd")

	vSwitch := &unstructured.Unstructured{}
	err := objUtils.Unmarshal(vSwitch, obj)
	if err != nil {
		flog.Warnf("unmarshal vSwitch data error: %v", obj)
	}

	status := objUtils.GetNestedString(vSwitch.Object,
		"spec", "status")
	if status == common.INIT {
		return
	}

	ip := objUtils.GetNestedString(vSwitch.Object,
		"spec", "ip")
	mask := objUtils.GetNestedString(vSwitch.Object,
		"spec", "mask")
	regionId := objUtils.GetNestedString(vSwitch.Object,
		"spec", "regionId")
	vpcId := objUtils.GetNestedString(vSwitch.Object,
		"spec", "vpcId")
	zone := objUtils.GetNestedString(vSwitch.Object,
		"spec", "zone")
	id := objUtils.GetNestedString(vSwitch.Object,
		"spec", "id")
	message := objUtils.GetNestedString(vSwitch.Object,
		"spec", "message")
	localName := objUtils.GetNestedString(vSwitch.Object,
		"spec", "localName")

	labels := map[string]interface{}{}
	getLabels := vSwitch.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	vSwitchObj := &networking.Vswitch{
		Metadata: core.Metadata{
			Name:      vSwitch.GetName(),
			Namespace: vSwitch.GetNamespace(),
			Workspace: getLabels["workspace"],
			Kind:      core.Kind(vSwitch.GetKind()),
			Labels:    labels,
		}, Spec: networking.VSwitchSpec{
			IP:        ip,
			Mask:      mask,
			Region:    regionId,
			Zone:      zone,
			Status:    status,
			VpcId:     vpcId,
			Id:        id,
			Message:   message,
			LocalName: localName,
		},
	}

	if vSwitchObj.Workspace == "" {
		vSwitchObj.Workspace = common.DefaultWorkspace
	}

	err = V.stage.Get(common.DefaultDatabase, common.VSWITCH, vSwitchObj.Name, &networking.Vswitch{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.VSWITCH, vSwitchObj)
		if createErr != nil {
			flog.Warnf("create vswitch err %v", createErr)
			return
		}
		flog.Infof("create vswitch %s, workspace: %s, namespace: %s", vSwitchObj.GetName(), vSwitchObj.GetWorkspace(), vSwitchObj.GetNamespace())
		return
	}
	if err != nil {
		flog.Warnf("get vswitch err %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.VSWITCH, vSwitchObj.Name, vSwitchObj, false)
	if err != nil {
		flog.Warnf("update vswitch err %v", err)
		return
	}
	if update {
		flog.Infof("update vSwitch %s ,workspace: %s, namespace: %s", vSwitchObj.GetName(), vSwitchObj.GetWorkspace(), vSwitchObj.GetNamespace())
	}

	return
}

func (V *VSwitchCtrl) SouthOnUpdate(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnUpdate")
	var force = false

	vSwitch := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(vSwitch, obj)
	if err != nil {
		flog.Warnf("unmarshal vSwitch data error: %v", obj)
		return
	}
	status := objUtils.GetNestedString(vSwitch.Object,
		"spec", "status")

	if status == common.DELETE || status == common.UPDATE {
		return
	}

	ip := objUtils.GetNestedString(vSwitch.Object,
		"spec", "ip")
	mask := objUtils.GetNestedString(vSwitch.Object,
		"spec", "mask")
	regionId := objUtils.GetNestedString(vSwitch.Object,
		"spec", "regionId")
	vpcId := objUtils.GetNestedString(vSwitch.Object,
		"spec", "vpcId")
	zone := objUtils.GetNestedString(vSwitch.Object,
		"spec", "zone")
	localName := objUtils.GetNestedString(vSwitch.Object,
		"spec", "localName")
	id := objUtils.GetNestedString(vSwitch.Object,
		"spec", "id")
	message := objUtils.GetNestedString(vSwitch.Object,
		"spec", "message")

	labels := map[string]interface{}{}
	getLabels := vSwitch.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	vSwitchObj := &networking.Vswitch{
		Metadata: core.Metadata{
			Name:      vSwitch.GetName(),
			Namespace: vSwitch.GetNamespace(),
			Workspace: getLabels["workspace"],
			Kind:      core.Kind(vSwitch.GetKind()),
			Labels:    labels,
		}, Spec: networking.VSwitchSpec{
			IP:        ip,
			Mask:      mask,
			Region:    regionId,
			Zone:      zone,
			Status:    status,
			VpcId:     vpcId,
			Id:        id,
			Message:   message,
			LocalName: localName,
		},
	}

	if vSwitchObj.Spec.Status == common.FAIL {
		force = true
		vSwitchObj.Metadata.IsDelete = false
	}

	if vSwitchObj.Workspace == "" {
		vSwitchObj.Workspace = common.DefaultWorkspace
	}

	_, _, err = V.stage.Apply(common.DefaultDatabase, common.VSWITCH, vSwitchObj.Name, vSwitchObj, force)
	if err != nil {
		flog.Warnf("update vSwitch error: %v", obj)
		return
	}
	flog.Infof("Apply a new vSwitch %s to stage", vSwitchObj.GetName())
	return
}

func (V *VSwitchCtrl) SouthOnDelete(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnDelete")

	vSwitch := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(vSwitch, obj)
	if err != nil {
		flog.Warnf("unmarshal vSwitch data error: %v", obj)
		return
	}

	labels := vSwitch.GetLabels()

	err = V.stage.Delete(common.DefaultDatabase, common.VSWITCH, vSwitch.GetName(), labels["workspace"])
	if err != nil {
		flog.Warnf("delete vSwitch error: %v", obj)
		return
	}

	flog.Infof("delete a vSwitch %s from stage", vSwitch.GetName())
	return
}

func (V *VSwitchCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
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

		watchInterface, err := client.Interface.Resource(vSwitchGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}
