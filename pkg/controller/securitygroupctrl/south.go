package securitygroupctrl

import (
	"context"
	"fmt"
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

func (V *SecurityGroupCtrl) SouthOnAdd(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnAdd")

	securityGroup := &unstructured.Unstructured{}
	err := objUtils.Unmarshal(securityGroup, obj)
	if err != nil {
		flog.Warnf("unmarshal securityGroup data error: %v", obj)
		return
	}

	isReturn, securityGroupObj := V.checkObjStatusAndGetObj(securityGroup, common.INIT)
	if isReturn {
		return
	}

	err = V.stage.Get(common.DefaultDatabase, common.SECURITYGROUP, securityGroupObj.Name, &networking.Vswitch{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.SECURITYGROUP, securityGroupObj)
		if createErr != nil {
			flog.Warnf("create securityGroup err %v", createErr)
			return
		}
		flog.Infof("create securityGroup %s, workspace: %s, namespace: %s", securityGroupObj.GetName(), securityGroupObj.GetWorkspace(), securityGroupObj.GetNamespace())
		return
	}
	if err != nil {
		flog.Warnf("get securityGroup err %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.SECURITYGROUP, securityGroupObj.Name, securityGroupObj, false)
	if err != nil {
		flog.Warnf("update securityGroup err %v", err)
		return
	}
	if update {
		flog.Infof("update securityGroup %s ,workspace: %s, namespace: %s", securityGroupObj.GetName(), securityGroupObj.GetWorkspace(), securityGroupObj.GetNamespace())
	}
	return
}

func (V *SecurityGroupCtrl) SouthOnUpdate(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnUpdate")
	var force = false

	securityGroup := &unstructured.Unstructured{}
	err := objUtils.Unmarshal(securityGroup, obj)
	if err != nil {
		flog.Warnf("unmarshal securityGroup data error: %v", obj)
		return
	}

	isReturn, securityGroupObj := V.checkObjStatusAndGetObj(securityGroup, common.DELETE, common.UPDATE)
	if isReturn {
		return
	}

	if securityGroupObj.Spec.Status == common.FAIL {
		force = true
		securityGroupObj.Metadata.IsDelete = false
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.SECURITYGROUP, securityGroupObj.Name, securityGroupObj, force)
	if err != nil {
		flog.Warnf("update securityGroupObj error: %v", obj)
		return
	}
	if update {
		flog.Infof("Apply a new securityGroupObj %s to stage", securityGroupObj.GetName())
	}
}

func (V *SecurityGroupCtrl) SouthOnDelete(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnDelete")

	securityGroup := &unstructured.Unstructured{}
	err := objUtils.Unmarshal(securityGroup, obj)
	if err != nil {
		flog.Warnf("unmarshal securityGroup data error: %v", obj)
		return
	}

	labels := securityGroup.GetLabels()

	err = V.stage.Delete(common.DefaultDatabase, common.SECURITYGROUP, securityGroup.GetName(), labels["workspace"])
	if err != nil {
		flog.Warnf("delete securityGroup error: %v", obj)
		return
	}

	flog.Infof("delete a securityGroup %s from stage", securityGroup.GetName())
	return
}

func (V *SecurityGroupCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
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

		watchInterface, err := client.Interface.Resource(securityGroupGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}

func (V *SecurityGroupCtrl) checkObjStatusAndGetObj(securityGroup *unstructured.Unstructured, checkStatus ...string) (bool, *system.SecurityGroup) {
	status := objUtils.GetNestedString(securityGroup.Object,
		"spec", "status")

	for _, s := range checkStatus {
		if status == s {
			return true, nil
		}
	}

	id := objUtils.GetNestedString(securityGroup.Object,
		"spec", "id")
	vpcId := objUtils.GetNestedString(securityGroup.Object,
		"spec", "vpcId")
	regionId := objUtils.GetNestedString(securityGroup.Object,
		"spec", "regionId")
	message := objUtils.GetNestedString(securityGroup.Object,
		"spec", "message")
	localName := objUtils.GetNestedString(securityGroup.Object,
		"spec", "localName")

	spec := objUtils.GetNestedMap(securityGroup.Object,
		"spec")

	labels := map[string]interface{}{}
	getLabels := securityGroup.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	securityGroupObj := &system.SecurityGroup{
		Metadata: core.Metadata{
			Name:      securityGroup.GetName(),
			Namespace: securityGroup.GetNamespace(),
			Workspace: getLabels["workspace"],
			Kind:      core.Kind(securityGroup.GetKind()),
			Labels:    labels,
		},
		Spec: system.SecurityGroupSpec{
			ID:        id,
			RegionId:  regionId,
			VpcId:     vpcId,
			Status:    status,
			Message:   message,
			LocalName: localName,
		},
	}

	ingressData := spec["ingress"]
	if ingressData != nil {
		for _, i2 := range ingressData.([]interface{}) {
			role := system.SecurityGroupRole{
				PortRange:    fmt.Sprintf("%v", i2.(map[string]interface{})["portRange"]),
				IpProtocol:   system.IpProtocolType(fmt.Sprintf("%v", i2.(map[string]interface{})["ipProtocol"])),
				SourceCidrIp: fmt.Sprintf("%v", i2.(map[string]interface{})["sourceCidrIp"]),
			}
			securityGroupObj.Spec.Ingress = append(securityGroupObj.Spec.Ingress, role)
		}
	}

	egressData := spec["egress"]
	if egressData != nil {
		for _, i2 := range egressData.([]interface{}) {
			role := system.SecurityGroupRole{
				PortRange:    fmt.Sprintf("%v", i2.(map[string]interface{})["portRange"]),
				IpProtocol:   system.IpProtocolType(fmt.Sprintf("%v", i2.(map[string]interface{})["ipProtocol"])),
				SourceCidrIp: fmt.Sprintf("%v", i2.(map[string]interface{})["sourceCidrIp"]),
			}
			securityGroupObj.Spec.Egress = append(securityGroupObj.Spec.Egress, role)
		}
	}

	if securityGroupObj.Workspace == "" {
		securityGroupObj.Workspace = common.DefaultWorkspace
	}
	return false, securityGroupObj
}
