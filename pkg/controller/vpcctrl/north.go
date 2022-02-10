package vpcctrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/networking"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (V *VPCCtrl) NorthOnAdd(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnAdd")

	vpc := networking.VirtualPrivateCloud{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &vpc); err != nil {
		flog.Infof("unstructured obj error %v", err)
		return
	}

	if vpc.Spec.Status != common.INIT {
		return
	}

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]vpcLabel, 0)
	for k, v := range vpc.Labels {
		label := vpcLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	virtualPrivateCloud := &virtualPrivateCloudParams{
		Name:      vpc.Name,
		Namespace: vpc.Namespace,
		Ip:        vpc.Spec.IP,
		Mask:      vpc.Spec.Mask,
		Region:    vpc.Spec.Region,
		Status:    vpc.Spec.Status,
		Labels:    labels,
		LocalName: vpc.Spec.LocalName,
	}

	unstructuredVirtualPrivateCloud, err := utilsObj.Render(virtualPrivateCloud, virtualPrivateCloudTpl)
	if err != nil {
		flog.Infof("render vpc obj error %v", err)
		return
	}
	_, err = client.Interface.Resource(virtualPrivateCloudGvr).Namespace(vpc.GetNamespace()).Create(
		context.Background(), unstructuredVirtualPrivateCloud, metav1.CreateOptions{})
	if err != nil {
		flog.Infof("create vpc error %v", err)
		return
	}

	flog.Infof("create a new vpc %s to k8s", vpc.GetName())
	return
}

func (V *VPCCtrl) NorthOnUpdate(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnUpdate")

	vpc := networking.VirtualPrivateCloud{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &vpc); err != nil {
		flog.Infof("unstructured vpc obj error %v", err)
		return
	}

	if vpc.Spec.Status != common.UPDATE {
		return
	}

	client, err := V.cs.GetClient(vpc.GetNamespace())
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]vpcLabel, 0)
	for k, v := range vpc.Labels {
		label := vpcLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	virtualPrivateCloud := &virtualPrivateCloudParams{
		Name:      vpc.Name,
		Namespace: vpc.Namespace,
		Ip:        vpc.Spec.IP,
		Mask:      vpc.Spec.Mask,
		Region:    vpc.Spec.Region,
		Status:    common.UPDATE,
		Id:        vpc.Spec.ID,
		Labels:    labels,
		LocalName: vpc.Spec.LocalName,
	}
	unstructuredVirtualPrivateCloud, err := utilsObj.Render(virtualPrivateCloud, virtualPrivateCloudTpl)
	if err != nil {
		flog.Infof("render vpc obj error %v", err)
		return
	}
	_, err = client.Interface.Resource(virtualPrivateCloudGvr).Namespace(vpc.GetNamespace()).Update(
		context.Background(), unstructuredVirtualPrivateCloud, metav1.UpdateOptions{})

	if err != nil {
		flog.Infof("update vpc obj error %v", err)
		return
	}

	flog.Infof("update vpc %s to k8s", vpc.GetName())
	return
}

func (V *VPCCtrl) NorthOnDelete(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnDelete")

	vpc := networking.VirtualPrivateCloud{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &vpc); err != nil {
		flog.Infof("unstructured obj error %v", err)
		return
	}
	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]vpcLabel, 0)
	for k, v := range vpc.Labels {
		label := vpcLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	virtualPrivateCloud := &virtualPrivateCloudParams{
		Name:      vpc.Name,
		Namespace: vpc.Namespace,
		Ip:        vpc.Spec.IP,
		Mask:      vpc.Spec.Mask,
		Region:    vpc.Spec.Region,
		Status:    common.DELETE,
		Id:        vpc.Spec.ID,
		Labels:    labels,
		LocalName: vpc.Spec.LocalName,
	}
	unstructuredVirtualPrivateCloud, err := utilsObj.Render(virtualPrivateCloud, virtualPrivateCloudTpl)
	if err != nil {
		flog.Infof("render vpc obj error %v", err)
		return
	}

	_, _, err = client.Apply(context.Background(), vpc.GetNamespace(), virtualPrivateCloudGvr, vpc.GetName(), unstructuredVirtualPrivateCloud, false)

	if err != nil {
		flog.Infof("delete vpc error %v", err)
		return
	}

	flog.Infof("trying delete vpc %s from k8s", vpc.GetName())
	return
}

func (V *VPCCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return V.stage.WatchEvent(ctx, common.DefaultDatabase, common.VPC, "0")
}
