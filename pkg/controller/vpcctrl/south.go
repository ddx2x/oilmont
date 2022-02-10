package vpcctrl

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

func (V *VPCCtrl) SouthOnAdd(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnAdd")
	vpc := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(vpc, obj)
	if err != nil {
		flog.Warnf("unmarshal vpc data error: %v", obj)
	}

	status := objUtils.GetNestedString(vpc.Object,
		"spec", "status")
	if status == common.INIT {
		return
	}

	ip := objUtils.GetNestedString(vpc.Object,
		"spec", "ip")
	mask := objUtils.GetNestedString(vpc.Object,
		"spec", "mask")
	regionId := objUtils.GetNestedString(vpc.Object,
		"spec", "regionId")
	vpcId := objUtils.GetNestedString(vpc.Object,
		"spec", "vpcId")
	message := objUtils.GetNestedString(vpc.Object,
		"spec", "message")
	localName := objUtils.GetNestedString(vpc.Object,
		"spec", "localName")

	labels := map[string]interface{}{}
	getLabels := vpc.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	vpcObj := &networking.VirtualPrivateCloud{
		Metadata: core.Metadata{
			Name:      vpc.GetName(),
			Namespace: vpc.GetNamespace(),
			Workspace: getLabels["workspace"],
			Kind:      core.Kind(vpc.GetKind()),
			Labels:    labels,
		}, Spec: networking.VirtualPrivateCloudSpec{
			IP:        ip,
			Mask:      mask,
			Region:    regionId,
			Status:    status,
			ID:        vpcId,
			Message:   message,
			LocalName: localName,
		},
	}

	if vpcObj.Workspace == "" {
		vpcObj.Workspace = common.DefaultWorkspace
	}

	err = V.stage.Get(common.DefaultDatabase, common.VPC, vpcObj.Name, &networking.VirtualPrivateCloud{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.VPC, vpcObj)
		if createErr != nil {
			flog.Warnf("create vpc error: %v", createErr)
			return
		}
		flog.Infof("create vpc %s,namespace: %s, workspace: %s", vpcObj.GetName(), vpcObj.GetNamespace(), vpcObj.GetWorkspace())
		return
	}
	if err != nil {
		flog.Warnf("get vpc error: %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.VPC, vpcObj.Name, vpcObj, false)
	if err != nil {
		flog.Warnf("create vpc error: %v", err)
		return
	}
	if update {
		flog.Infof("update vpc %s, workspace: %s, namespace: %s", vpcObj.GetName(), vpcObj.GetWorkspace(), vpcObj.GetNamespace())
	}
	return
}

func (V *VPCCtrl) SouthOnUpdate(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnUpdate")
	var force = false
	vpc := &unstructured.Unstructured{}
	err := objUtils.Unmarshal(vpc, obj)
	if err != nil {
		flog.Warnf("unmarshal vpc data error: %v", obj)
	}
	status := objUtils.GetNestedString(vpc.Object,
		"spec", "status")
	if status == common.DELETE || status == common.UPDATE {
		return
	}

	ip := objUtils.GetNestedString(vpc.Object,
		"spec", "ip")
	mask := objUtils.GetNestedString(vpc.Object,
		"spec", "mask")
	regionId := objUtils.GetNestedString(vpc.Object,
		"spec", "regionId")
	vpcId := objUtils.GetNestedString(vpc.Object,
		"spec", "vpcId")
	message := objUtils.GetNestedString(vpc.Object,
		"spec", "message")
	localName := objUtils.GetNestedString(vpc.Object,
		"spec", "localName")

	labels := map[string]interface{}{}
	getLabels := vpc.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	vpcObj := &networking.VirtualPrivateCloud{
		Metadata: core.Metadata{
			Name:      vpc.GetName(),
			Namespace: vpc.GetNamespace(),
			Workspace: getLabels["workspace"],
			Kind:      core.Kind(vpc.GetKind()),
			Labels:    labels,
		}, Spec: networking.VirtualPrivateCloudSpec{
			IP:        ip,
			Mask:      mask,
			Region:    regionId,
			Status:    status,
			ID:        vpcId,
			Message:   message,
			LocalName: localName,
		},
	}

	if vpcObj.Spec.Status == common.FAIL {
		force = true
		vpcObj.Metadata.IsDelete = false
	}

	if vpcObj.Workspace == "" {
		vpcObj.Workspace = common.DefaultWorkspace
	}

	_, _, err = V.stage.Apply(common.DefaultDatabase, common.VPC, vpcObj.Name, vpcObj, force)
	if err != nil {
		flog.Warnf("update vpc error: %v", obj)
		return
	}

	flog.Infof("update a vpc %s to stage", vpcObj.GetName())
	return
}

func (V *VPCCtrl) SouthOnDelete(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnDelete")
	vpc := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(vpc, obj)
	if err != nil {
		flog.Warnf("unmarshal vpc data error: %v", obj)
	}

	labels := vpc.GetLabels()

	err = V.stage.Delete(common.DefaultDatabase, common.VPC, vpc.GetName(), labels["workspace"])
	if err != nil {
		flog.Warnf("delete vpc error: %v", obj)
		return
	}

	flog.Warnf("delete a vpc %s from stage", vpc.GetName())
	return
}

func (V *VPCCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
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

		watchInterface, err := client.Interface.Resource(virtualPrivateCloudGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}
