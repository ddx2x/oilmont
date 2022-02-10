package securitygroupctrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (V *SecurityGroupCtrl) NorthOnAdd(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnAdd")

	securityGroup := system.SecurityGroup{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &securityGroup); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	if securityGroup.Spec.Status != common.INIT {
		return
	}

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]securityGroupLabel, 0)
	for k, v := range securityGroup.Labels {
		label := securityGroupLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	sg := securityGroupParams{
		Name:      securityGroup.GetName(),
		Namespace: securityGroup.GetNamespace(),
		Id:        securityGroup.Spec.ID,
		VpcId:     securityGroup.Spec.VpcId,
		Region:    securityGroup.Spec.RegionId,
		Ingress:   securityGroup.Spec.Ingress,
		Egress:    securityGroup.Spec.Egress,
		LocalName: securityGroup.Spec.LocalName,
		Status:    common.INIT,
		Labels:    labels,
	}

	unstructuredObj, err := utilsObj.Render(sg, securityGroupTpl)
	if err != nil {
		flog.Warnf("render obj error %v", err)
		return
	}
	_, err = client.Interface.Resource(securityGroupGvr).Namespace(securityGroup.GetNamespace()).Create(
		context.Background(), unstructuredObj, metav1.CreateOptions{})

	if err != nil {
		flog.Warnf("create securityGroup error %v", err)
		return
	}

	flog.Infof("create securityGroup %s to k8s", securityGroup.GetName())
	return

}

func (V *SecurityGroupCtrl) NorthOnUpdate(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnUpdate")

	securityGroup := system.SecurityGroup{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &securityGroup); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	if securityGroup.Spec.Status != common.UPDATE {
		return
	}

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]securityGroupLabel, 0)
	for k, v := range securityGroup.Labels {
		label := securityGroupLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	sg := securityGroupParams{
		Name:      securityGroup.GetName(),
		Namespace: securityGroup.GetNamespace(),
		Id:        securityGroup.Spec.ID,
		VpcId:     securityGroup.Spec.VpcId,
		Region:    securityGroup.Spec.RegionId,
		Ingress:   securityGroup.Spec.Ingress,
		Egress:    securityGroup.Spec.Egress,
		Status:    common.UPDATE,
		Labels:    labels,
		LocalName: securityGroup.Spec.LocalName,
	}

	unstructuredObj, err := utilsObj.Render(sg, securityGroupTpl)
	if err != nil {
		flog.Warnf("render obj error %v", err)
		return
	}

	_, _, err = client.Apply(context.Background(), securityGroup.GetNamespace(), securityGroupGvr, securityGroup.GetName(), unstructuredObj, false)

	if err != nil {
		flog.Infof("update securityGroup obj error %v", err)
		return
	}

	flog.Infof("update securityGroup %s to k8s", securityGroup.GetName())
	return
}

func (V *SecurityGroupCtrl) NorthOnDelete(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnUpdate")

	securityGroup := system.SecurityGroup{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, &securityGroup); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Infof("get client error %v", err)
		return
	}

	labels := make([]securityGroupLabel, 0)
	for k, v := range securityGroup.Labels {
		label := securityGroupLabel{
			Key:   k,
			Value: v,
		}
		labels = append(labels, label)
	}

	if securityGroup.Spec.Ingress == nil {
		securityGroup.Spec.Ingress = make([]system.SecurityGroupRole, 0)
	}
	if securityGroup.Spec.Egress == nil {
		securityGroup.Spec.Egress = make([]system.SecurityGroupRole, 0)
	}

	sg := securityGroupParams{
		Name:      securityGroup.GetName(),
		Namespace: securityGroup.GetNamespace(),
		Id:        securityGroup.Spec.ID,
		VpcId:     securityGroup.Spec.VpcId,
		Region:    securityGroup.Spec.RegionId,
		Ingress:   securityGroup.Spec.Ingress,
		Egress:    securityGroup.Spec.Egress,
		Status:    common.DELETE,
		Labels:    labels,
		LocalName: securityGroup.Spec.LocalName,
	}

	unstructuredObj, err := utilsObj.Render(sg, securityGroupTpl)
	if err != nil {
		flog.Warnf("render obj error %v", err)
		return
	}

	_, _, err = client.Apply(context.Background(), securityGroup.GetNamespace(), securityGroupGvr, securityGroup.GetName(), unstructuredObj, false)

	if err != nil {
		flog.Infof("delete securityGroup obj error %v", err)
		return
	}

	flog.Infof("delete securityGroup %s to k8s", securityGroup.GetName())
	return
}

func (V *SecurityGroupCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return V.stage.WatchEvent(ctx, common.DefaultDatabase, common.SECURITYGROUP, "0")
}
