package networkinterfacectrl

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

func (V *NetworkInterfaceCtrl) SouthOnAdd(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnAdd")
	unstructuredENI := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(unstructuredENI, obj)
	if err != nil {
		flog.Warnf("unmarshal networkInterface data error: %v", obj)
	}

	networkInterface, ok := V.checkObjStatusAndGetObj(unstructuredENI, common.INIT)
	if !ok {
		return
	}

	err = V.stage.Get(common.DefaultDatabase, common.NETWORKINTERFACE, networkInterface.GetName(), &networking.NetworkInterface{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.NETWORKINTERFACE, networkInterface)
		if createErr != nil {
			flog.Warnf("create networkInterface error: %v", createErr)
			return
		}
		flog.Infof("create networkInterface %s,namespace: %s, workspace: %s", networkInterface.GetName(), networkInterface.GetNamespace(), networkInterface.GetWorkspace())
		return
	}
	if err != nil {
		flog.Warnf("get networkInterface error: %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.NETWORKINTERFACE, networkInterface.GetName(), networkInterface, false)
	if err != nil {
		flog.Warnf("create networkInterface error: %v", err)
		return
	}
	if update {
		flog.Infof("update networkInterface %s, workspace: %s, namespace: %s", networkInterface.GetName(), networkInterface.GetWorkspace(), networkInterface.GetNamespace())
	}
	return
}

func (V *NetworkInterfaceCtrl) SouthOnUpdate(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnUpdate")
	var force = false
	unstructuredENI := &unstructured.Unstructured{}
	err := objUtils.Unmarshal(unstructuredENI, obj)
	if err != nil {
		flog.Warnf("unmarshal networkInterface data error: %v", obj)
	}

	networkInterface, ok := V.checkObjStatusAndGetObj(unstructuredENI, common.DELETE, common.UPDATE)
	if !ok {
		return
	}

	if networkInterface.Spec.Status == common.FAIL {
		force = true
		networkInterface.Metadata.IsDelete = false
	}

	_, _, err = V.stage.Apply(common.DefaultDatabase, common.NETWORKINTERFACE, networkInterface.GetName(), networkInterface, force)
	if err != nil {
		flog.Warnf("update networkInterface error: %v", obj)
		return
	}

	flog.Infof("update a networkInterface %s to stage", networkInterface.GetName())
	return
}

func (V *NetworkInterfaceCtrl) SouthOnDelete(obj runtime.Object) {
	flog := V.flog.WithField("func", "SouthOnDelete")
	unstructuredENI := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(unstructuredENI, obj)
	if err != nil {
		flog.Warnf("unmarshal networkInterface data error: %v", obj)
	}

	labels := unstructuredENI.GetLabels()

	err = V.stage.Delete(common.DefaultDatabase, common.NETWORKINTERFACE, unstructuredENI.GetName(), labels["workspace"])
	if err != nil {
		flog.Warnf("delete networkInterface error: %v", obj)
		return
	}

	flog.Warnf("delete a networkInterface %s from stage", unstructuredENI.GetName())
	return
}

func (V *NetworkInterfaceCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
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

		watchInterface, err := client.Interface.Resource(NetworkInterfaceGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}

func (V *NetworkInterfaceCtrl) checkObjStatusAndGetObj(unstructuredENI *unstructured.Unstructured, checkStatus ...string) (*networking.NetworkInterface, bool) {
	status := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "status")
	for _, s := range checkStatus {
		if status == s {
			return nil, false
		}
	}

	description := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "description")
	localName := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "localName")
	macAddress := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "macAddress")
	message := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "message")
	networkInterfaceId := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "networkInterfaceId")
	privateDnsName := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "privateDnsName")
	privateIpAddress := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "privateIpAddress")
	publicDnsName := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "publicDnsName")
	publicIpAddress := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "publicIpAddress")
	region := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "region")
	state := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "state")
	subnetId := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "subnetId")
	eniType := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "type")
	vpcId := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "vpcId")
	zone := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "zone")
	attachTime := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "attachment", "attachTime")
	instanceId := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "attachment", "instanceId")
	attachStatus := objUtils.GetNestedString(unstructuredENI.Object,
		"spec", "attachment", "status")

	spec := objUtils.GetNestedMap(unstructuredENI.Object,
		"spec")

	labels := map[string]interface{}{}
	getLabels := unstructuredENI.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	networkInterface := &networking.NetworkInterface{
		Metadata: core.Metadata{
			Name:      unstructuredENI.GetName(),
			Namespace: unstructuredENI.GetNamespace(),
			Workspace: getLabels["workspace"],
			Kind:      core.Kind(unstructuredENI.GetKind()),
			Labels:    labels,
		},
		Spec: networking.NetworkInterfaceSpec{
			Description:      description,
			PrivateDnsName:   privateDnsName,
			PrivateIpAddress: privateIpAddress,
			PublicDnsName:    publicDnsName,
			PublicIpAddress:  publicIpAddress,
			SubnetId:         subnetId,
			VPCId:            vpcId,
			MacAddress:       macAddress,
			Type:             eniType,
			ID:               networkInterfaceId,
			LocalName:        localName,
			Region:           region,
			Zone:             zone,
			Status:           status,
			State:            state,
			Message:          message,
			Attachment: networking.ENIAttachment{
				AttachTime: attachTime,
				InstanceId: instanceId,
				Status:     attachStatus,
			},
		},
	}
	privateIpSets := spec["privateIpSets"]
	if privateIpSets != nil {
		for _, obj := range privateIpSets.([]interface{}) {
			_privateIpSet := obj.(map[string]interface{})
			ipset := networking.ENIPrivateIpSet{
				Primary:          _privateIpSet["primary"].(bool),
				PrivateDnsName:   _privateIpSet["privateDnsName"].(string),
				PrivateIpAddress: _privateIpSet["privateIpAddress"].(string),
				PublicIpAddress:  _privateIpSet["publicIpAddress"].(string),
			}
			networkInterface.Spec.PrivateIpSets = append(networkInterface.Spec.PrivateIpSets, ipset)
		}
	}
	securityGroupIds := spec["securityGroupIds"]
	if securityGroupIds != nil {
		for _, _obj := range securityGroupIds.([]interface{}) {
			networkInterface.Spec.SecurityGroupIds = append(networkInterface.Spec.SecurityGroupIds, _obj.(string))
		}
	}
	ipv6 := spec["ipv6"]
	if ipv6 != nil {
		for _, _obj := range ipv6.([]interface{}) {
			networkInterface.Spec.Ipv6 = append(networkInterface.Spec.Ipv6, _obj.(string))
		}
	}

	if networkInterface.Workspace == "" {
		networkInterface.Workspace = common.DefaultWorkspace
	}
	return networkInterface, true
}
