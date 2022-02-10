package vswitchctrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/networking"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (V *VSwitchCtrl) NorthOnAdd(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnAdd")

	vSwitch := networking.Vswitch{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &vSwitch); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	if vSwitch.Spec.Status != common.INIT {
		return
	}

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]vpcLabel, 0)
	for k, v := range vSwitch.Labels {
		label := vpcLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	virtualPrivateCloud := &virtualPrivateCloudParams{
		Name:      vSwitch.Name,
		Namespace: vSwitch.Namespace,
		Ip:        vSwitch.Spec.IP,
		Mask:      vSwitch.Spec.Mask,
		Region:    vSwitch.Spec.Region,
		Status:    common.INIT,
		VpcId:     vSwitch.Spec.VpcId,
		Zone:      vSwitch.Spec.Zone,
		Labels:    labels,
		LocalName: vSwitch.Spec.LocalName,
	}

	unstructuredVirtualPrivateCloud, err := utilsObj.Render(virtualPrivateCloud, vSwitchTpl)
	if err != nil {
		flog.Warnf("render obj error %v", err)
		return
	}
	_, err = client.Interface.Resource(vSwitchGvr).Namespace(vSwitch.GetNamespace()).Create(
		context.Background(), unstructuredVirtualPrivateCloud, metav1.CreateOptions{})

	if err != nil {
		flog.Warnf("create vSwitch error %v", err)
		return
	}

	flog.Infof("create vSwitch %s to k8s", vSwitch.GetName())
	return
}

func (V *VSwitchCtrl) NorthOnUpdate(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnUpdate")

	vSwitch := networking.Vswitch{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &vSwitch); err != nil {
		flog.Infof("unstructured obj error %v", err)
		return
	}

	if vSwitch.Spec.Status != common.UPDATE {
		return
	}

	client, err := V.cs.GetClient(vSwitch.GetNamespace())
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]vpcLabel, 0)
	for k, v := range vSwitch.Labels {
		label := vpcLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	virtualPrivateCloud := &virtualPrivateCloudParams{
		Name:      vSwitch.Name,
		Namespace: vSwitch.Namespace,
		Ip:        vSwitch.Spec.IP,
		Mask:      vSwitch.Spec.Mask,
		Region:    vSwitch.Spec.Region,
		Status:    common.UPDATE,
		VpcId:     vSwitch.Spec.VpcId,
		Zone:      vSwitch.Spec.Zone,
		Id:        vSwitch.Spec.Id,
		Labels:    labels,
		LocalName: vSwitch.Spec.LocalName,
	}

	unstructuredVirtualPrivateCloud, err := utilsObj.Render(virtualPrivateCloud, vSwitchTpl)
	if err != nil {
		flog.Infof("render vSwitch obj error %v", err)
		return
	}

	_, _, err = client.Apply(context.Background(), vSwitch.GetNamespace(), vSwitchGvr, vSwitch.GetName(), unstructuredVirtualPrivateCloud, false)

	if err != nil {
		flog.Infof("update vSwitch obj error %v", err)
		return
	}

	flog.Infof("update vSwitch %s to k8s", vSwitch.GetName())
	return
}

func (V *VSwitchCtrl) NorthOnDelete(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnDelete")

	vSwitch := networking.Vswitch{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &vSwitch); err != nil {
		flog.Infof("unstructured obj error %v", err)
		return
	}
	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]vpcLabel, 0)
	for k, v := range vSwitch.Labels {
		label := vpcLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	virtualPrivateCloud := &virtualPrivateCloudParams{
		Name:      vSwitch.Name,
		Namespace: vSwitch.Namespace,
		Id:        vSwitch.Spec.Id,
		Ip:        vSwitch.Spec.IP,
		Mask:      vSwitch.Spec.Mask,
		Region:    vSwitch.Spec.Region,
		Status:    common.DELETE,
		VpcId:     vSwitch.Spec.VpcId,
		Zone:      vSwitch.Spec.Zone,
		Labels:    labels,
		LocalName: vSwitch.Spec.LocalName,
	}

	unstructuredVirtualPrivateCloud, err := utilsObj.Render(virtualPrivateCloud, vSwitchTpl)
	if err != nil {
		flog.Infof("render vSwitch obj error %v", err)
		return
	}

	_, _, err = client.Apply(context.Background(), vSwitch.GetNamespace(), vSwitchGvr, vSwitch.GetName(), unstructuredVirtualPrivateCloud, false)

	if err != nil {
		flog.Infof("delete vSwitch obj error %v", err)
		return
	}

	flog.Infof("delete vSwitch %s to k8s", vSwitch.GetName())
	return
}

func (V *VSwitchCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return V.stage.WatchEvent(ctx, common.DefaultDatabase, common.VSWITCH, "0")
}
