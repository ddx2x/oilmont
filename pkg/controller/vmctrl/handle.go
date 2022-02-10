package vmctrl

import (
	"context"
	"fmt"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/controller/clients"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/compute"
	"github.com/ddx2x/oilmont/pkg/resource/networking"
	objUtils "github.com/ddx2x/oilmont/pkg/utils/obj"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kubeVirtV1 "kubevirt.io/client-go/apis/core/v1"
	"kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"
)

func (V *VMCtrl) addThirdPartyToK8s(vm *compute.VirtualMachine) {
	flog := V.flog.WithField("func", "addThirdPartyToK8s")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Warnf("get client error %s", common.DefaultKubernetes)
		return
	}

	unstructuredVirtualMachine, err := V.getUnstructuredObj(vm)
	if err != nil {
		flog.Warnf("render VirtualMachine error %v", err)
		return
	}

	_, err = client.Interface.Resource(virtualMachineGvr).Namespace(vm.GetNamespace()).Create(
		context.Background(), unstructuredVirtualMachine, metav1.CreateOptions{})
	if err != nil {
		flog.Warnf("create vm error %v", err)
		return
	}

	flog.Infof("create vm %s complete", vm.GetName())
}

func (V *VMCtrl) updateThirdPartyToK8s(vm *compute.VirtualMachine) {
	flog := V.flog.WithField("func", "updateThirdPartyToK8s")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Warnf("get client error %s", common.DefaultKubernetes)
		return
	}

	unstructuredVirtualMachine, err := V.getUnstructuredObj(vm)
	if err != nil {
		flog.Warnf("render VirtualMachine error %v", err)
		return
	}

	if _, _, err := client.Apply(context.Background(), vm.GetNamespace(), virtualMachineGvr, vm.GetName(), unstructuredVirtualMachine, false); err != nil {
		flog.Warnf("apply VirtualMachine error %v", err)
		return
	}
	flog.Infof("apply vm %s complete", vm.GetName())
}

func (V *VMCtrl) deleteK8sDataThirdPartyVM(vm *compute.VirtualMachine) {
	flog := V.flog.WithField("func", "deleteK8sDataThirdPartyVM")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		return
	}

	unstructuredVirtualMachine, err := V.getUnstructuredObj(vm)
	if err != nil {
		flog.Warnf("render VirtualMachine error %v", err)
		return
	}

	_, getErr := client.Interface.Resource(virtualMachineGvr).Namespace(vm.GetNamespace()).Get(
		context.Background(), vm.GetName(), metav1.GetOptions{})
	if errors.IsNotFound(getErr) {
		return
	}

	_, _, err = client.Apply(context.Background(), vm.GetNamespace(), virtualMachineGvr, vm.GetName(), unstructuredVirtualMachine, false)

	if err != nil {
		flog.Infof("delete vm obj error %v", err)
		return
	}

	flog.Infof("trying delete vm %s to k8s", vm.GetName())
}

func (V *VMCtrl) getUnstructuredObj(vm *compute.VirtualMachine) (*unstructured.Unstructured, error) {
	virtualMachine := &thirdVirtualMachineParams{
		Name:          vm.GetName(),
		Region:        vm.Spec.RegionId,
		Zone:          vm.Spec.Az,
		SecurityGroup: vm.Spec.SecurityGroup,
		Vendor:        vm.Spec.Vendor,
		Image:         vm.Spec.ImageId,
		VpcId:         vm.Spec.VPCId,
		VSwitchId:     vm.Spec.VSwitchId,
		InstanceType:  vm.Spec.InstanceType,
		Id:            vm.Spec.InstanceId,
		Status:        vm.Spec.Status,
		Namespace:     vm.GetNamespace(),
		RootDevice:    vm.Spec.RootDevice,
		LocalName:     vm.Spec.LocalName,
		State:         string(vm.Spec.State),
		CPU:           vm.Spec.CPU,
		Memory:        vm.Spec.Memory,
		CreateTime:    vm.Spec.CreateTime,
		Os:            vm.Spec.Os,
	}

	for k, v := range vm.Labels {
		labels := vmLabel{
			Key:   k,
			Value: v,
		}
		virtualMachine.Labels = append(virtualMachine.Labels, labels)
	}
	return objUtils.Render(virtualMachine, thirdVirtualMachineTpl)
}

func (V *VMCtrl) addThirdVmToStage(obj runtime.Object) {
	flog := V.flog.WithField("func", "addThirdVmToStage")
	var update = true

	virtualMachine := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(virtualMachine, obj)
	if err != nil {
		flog.Warnf("unmarshal virtualMachine data error: %v", obj)
	}

	ok, vm := V.checkObjStatusAndGetObj(virtualMachine, common.INIT)
	if !ok {
		return
	}

	vmFilter := map[string]interface{}{"metadata.name": vm.GetName(), "metadata.workspace": vm.GetWorkspace()}
	err = V.stage.GetByFilter(common.DefaultDatabase, common.VIRTUALMACHINE, &compute.VirtualMachine{}, vmFilter, false)

	if err == datasource.NotFound {
		update = false
	}
	if err != datasource.NotFound && err != nil {
		flog.Warnf("get data error: %v", err)
		return
	}

	if update {
		_, updated, err := V.stage.Apply(common.DefaultDatabase, common.VIRTUALMACHINE, vm.Name, vm, false)
		if err != nil {
			flog.Warnf("virtualMachine data update to mongo error: %v", obj)
			return
		}
		if updated {
			flog.Infof("virtualMachine data update to mongo %s", vm.Name)
		}
	} else {
		_, err = V.stage.Create(common.DefaultDatabase, common.VIRTUALMACHINE, vm)
		if err != nil {
			flog.Warnf("virtualMachine data save to mongo error: %v", obj)
			return
		}
		flog.Infof("virtualMachine data save to mongo %s", vm.Name)
	}
}

func (V *VMCtrl) updateThirdVmToStage(obj runtime.Object) {
	flog := V.flog.WithField("func", "updateThirdVmToStage")
	var force = false
	virtualMachine := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(virtualMachine, obj)
	if err != nil {
		flog.Warnf("unmarshal virtualMachine data error: %v", obj)
	}

	ok, vm := V.checkObjStatusAndGetObj(virtualMachine, common.DELETE, common.UPDATE)
	if !ok {
		return
	}

	if vm.Spec.Status == common.FAIL {
		force = true
		vm.Metadata.IsDelete = false
	}

	// 两个字段是 list, monogo上会出现无法清空的情况
	//storages := vm.Spec.Storage
	//eni := vm.Spec.NetWorkInterface
	//vm.Spec.Storage = nil
	//vm.Spec.NetWorkInterface = nil
	//paths := []string{
	//	"spec.storage",
	//	"spec.network_interface",
	//}
	//V.stage.Apply(common.DefaultDatabase, common.VIRTUALMACHINE, vm.Name, vm, false, paths...)
	//vm.Spec.Storage = storages
	//vm.Spec.NetWorkInterface = eni

	_, updated, err := V.stage.Apply(common.DefaultDatabase, common.VIRTUALMACHINE, vm.Name, vm, force)
	if err != nil {
		flog.Warnf("update vm err %v", err)
	}
	if updated {
		flog.Infof("update vm %s, workspace: %s, namespace: %s", vm.GetName(), vm.GetWorkspace(), vm.GetNamespace())
	}
}

func (V *VMCtrl) applyLiZiVmToK8s(vm *compute.VirtualMachine) {
	flog := V.flog.WithField("func", "applyLiZiVmToK8s")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		flog.Warnf("get client error %s", vm.Spec.Vendor)
		return
	}

	for _, obj := range vm.Spec.Storage {
		storage := obj
		if err = V.CreateOrApplyDataVolume(ctx, client, vm.GetWorkspace(), &storage); err != nil {
			flog.Warnf("%v", err)
			if applyErr := V.changeVMStatus(vm, common.FAIL, err.Error()); applyErr != nil {
				flog.Warnf("change vm status error %v", applyErr)
			}
			return
		}
		flog.Infof("create dataVolume %s to k8s ", storage.Name)
	}

	if err = V.CreateOrApplyVMI(client, vm); err != nil {
		flog.Warnf("%v", err)
		if applyErr := V.changeVMStatus(vm, common.FAIL, err.Error()); applyErr != nil {
			flog.Warnf("change vm status error %v", applyErr)
		}
		return
	}

	flog.Infof("create vm %s to k8s ", vm.GetName())
	if applyErr := V.changeVMStatus(vm, common.RUNNING, "success"); applyErr != nil {
		flog.Warnf("change vm status error %v", applyErr)
	}
}

func (V *VMCtrl) deleteLiZiVM(vm *compute.VirtualMachine) {
	flog := V.flog.WithField("func", "deleteLiZiVM")

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	if err != nil {
		return
	}

	if err := client.Interface.Resource(virtualMachineInstanceGvr).Delete(context.Background(), vm.GetName(), metav1.DeleteOptions{}); err != nil {
		flog.Warnf("delete vm %s error %v", vm.GetName(), err)
		return
	}

	dataVolumeName := fmt.Sprintf("dataVolume-%s", vm.GetName())

	err = client.Interface.Resource(dataVolumeGvr).Delete(context.Background(), dataVolumeName, metav1.DeleteOptions{})
	if err != nil {
		flog.Warnf("delete dataVolume %s error %v", dataVolumeName, err)
	}
}

func (V *VMCtrl) applyLiZiVmToStage(obj runtime.Object) {
	flog := V.flog.WithField("func", "applyLiZiVmToStage")

	var update = true
	virtualMachine := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(virtualMachine, obj)
	if err != nil {
		flog.Warnf("unmarshal virtualMachine data error: %v", obj)
	}

	getObj := &compute.VirtualMachine{}
	err = V.stage.Get(common.DefaultDatabase, common.VIRTUALMACHINE, virtualMachine.GetName(), getObj, false)
	if err == datasource.NotFound {
		update = false
	}

	client, err := V.cs.GetClient(common.DefaultKubernetes)
	vmi, err := client.KubevirtCli.VirtualMachineInstance(getObj.GetWorkspace()).Get(getObj.GetName(), &metav1.GetOptions{})
	if errors.IsNotFound(err) {
		update = false
	}

	if err != datasource.NotFound && err != nil {
		flog.Warnf("get data error: %v", err)
		return
	}

	region := vmi.Labels["cloud.ddx2x.nip/region"]
	zone := vmi.Labels["cloud.ddx2x.nip/availability-zone"]
	vendor := vmi.Labels["cloud.ddx2x.nip/vendor"]

	state := vmi.Status.Phase
	message := vmi.Status.Reason

	vm := &compute.VirtualMachine{
		Metadata: core.Metadata{
			Name:      virtualMachine.GetName(),
			Kind:      core.Kind(virtualMachine.GetKind()),
			Namespace: virtualMachine.GetNamespace(),
		},
		Spec: compute.VirtualMachineSpec{
			Vendor:   vendor,
			Az:       zone,
			RegionId: region,
			State:    compute.VirtualMachineStateType(state),
			Message:  message,
			Status:   common.RUNNING,
		},
	}

	for _, volume := range vmi.Spec.Volumes {
		if volume.DataVolume == nil {
			continue
		}
		volumeName := volume.DataVolume.Name

		dataVolume, err := client.KubevirtCli.CdiClient().CdiV1beta1().
			DataVolumes(getObj.GetWorkspace()).
			Get(context.Background(), volumeName, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			update = false
		}
		if err != datasource.NotFound && err != nil {
			flog.Warnf("get dataVolume error: %v", err)
			return
		}

		storage := compute.VmStorage{
			Name:     dataVolume.GetName(),
			Quantity: dataVolume.Spec.PVC.Resources.Requests.Storage().String(),
			Registry: dataVolume.Spec.Source.Registry.URL,
			Status:   string(dataVolume.Status.Phase),
			Type:     "DataVolume",
		}
		vm.Spec.Storage = append(vm.Spec.Storage, storage)
	}

	if update {
		_, _, err = V.stage.Apply(common.DefaultDatabase, common.VIRTUALMACHINE, vm.Name, vm, false)
		if err != nil {
			flog.Warnf("virtualMachine data update to mongo error: %v", obj)
		}
	} else {
		_, err = V.stage.Create(common.DefaultDatabase, common.VIRTUALMACHINE, vm)
		if err != nil {
			flog.Warnf("virtualMachine data save to mongo error: %v", obj)
		}
	}
}

func (V *VMCtrl) CreateOrApplyDataVolume(ctx context.Context, client *clients.KubeClient, workspace string, storage *compute.VmStorage) error {
	var update = true
	dataVolume := &v1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataVolume",
			APIVersion: "cdi.kubevirt.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      storage.Name,
			Namespace: workspace,
		},
		Spec: v1beta1.DataVolumeSpec{
			Source: &v1beta1.DataVolumeSource{
				Registry: &v1beta1.DataVolumeSourceRegistry{
					//TODO: 自定义
					URL: "docker://laiks/fedora:cloud-base",
				},
			},

			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(storage.Quantity),
					},
				},
			},
		},
	}

	_, err := client.KubevirtCli.CdiClient().CdiV1beta1().
		DataVolumes(workspace).
		Get(ctx, dataVolume.GetName(), metav1.GetOptions{})

	if errors.IsNotFound(err) {
		update = false
	}
	if !errors.IsNotFound(err) && err != nil {
		return fmt.Errorf("get dataVolume error %v", err)
	}

	if update {
		_, _, err = client.ApplyDataVolume(ctx, dataVolume.GetNamespace(), dataVolume.GetName(), dataVolume, false)
		if err != nil {
			return fmt.Errorf("update dataVolume %s error %v", storage.Name, err)
		}
	} else {
		_, err = client.KubevirtCli.
			CdiClient().CdiV1beta1().
			DataVolumes(workspace).
			Create(ctx, dataVolume, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("create dataVolume %s error %v", storage.Name, err)
		}
	}
	return nil
}

func (V *VMCtrl) CreateOrApplyVMI(client *clients.KubeClient, vm *compute.VirtualMachine) error {
	var update = true
	cloudInitDisk := fmt.Sprintf("cloudinitdisk-%s", vm.Name)

	vmiResource := corev1.ResourceList{}
	vmiResource[corev1.ResourceMemory] = resource.MustParse("1Gi")

	vmi := &kubeVirtV1.VirtualMachineInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachineInstance",
			APIVersion: "kubevirt.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      vm.GetName(),
			Namespace: vm.GetWorkspace(),
			Labels: map[string]string{
				"kubevirt.io/vm":                   vm.GetName(),
				"cloud.ddx2x.nip/region":            vm.Spec.RegionId,
				"cloud.ddx2x.nip/availability-zone": vm.Spec.Az,
				"cloud.ddx2x.nip/vendor":            vm.Spec.Vendor,
			},
		},
		Spec: kubeVirtV1.VirtualMachineInstanceSpec{
			Domain: kubeVirtV1.DomainSpec{
				Devices: kubeVirtV1.Devices{
					Disks: []kubeVirtV1.Disk{
						{
							Name:       cloudInitDisk,
							DiskDevice: kubeVirtV1.DiskDevice{Disk: &kubeVirtV1.DiskTarget{Bus: "virtio"}},
						},
					},
				},
				Resources: kubeVirtV1.ResourceRequirements{
					Requests: vmiResource,
				},
			},
			Volumes: []kubeVirtV1.Volume{
				{
					Name: cloudInitDisk,
					VolumeSource: kubeVirtV1.VolumeSource{
						CloudInitNoCloud: &kubeVirtV1.CloudInitNoCloudSource{
							UserData: fmt.Sprintf("#cloud-config\npassword: 12345\nchpasswd: \n expire: False\n list:\n \\ - root:12345"),
						},
					},
				},
			},
		},
	}

	for _, storage := range vm.Spec.Storage {
		dataVolumeDisk := fmt.Sprintf("datavolumedisk-%s", storage.Name)
		vmi.Spec.Domain.Devices.Disks = append(vmi.Spec.Domain.Devices.Disks, kubeVirtV1.Disk{Name: dataVolumeDisk, DiskDevice: kubeVirtV1.DiskDevice{Disk: &kubeVirtV1.DiskTarget{Bus: "virtio"}}})
		vmi.Spec.Volumes = append(vmi.Spec.Volumes, kubeVirtV1.Volume{
			Name:         dataVolumeDisk,
			VolumeSource: kubeVirtV1.VolumeSource{DataVolume: &kubeVirtV1.DataVolumeSource{Name: storage.Name}}})
	}

	_, err := client.KubevirtCli.VirtualMachineInstance(vm.GetWorkspace()).Get(vm.GetName(), &metav1.GetOptions{})
	if errors.IsNotFound(err) {
		update = false
	}
	if !errors.IsNotFound(err) && err != nil {
		return fmt.Errorf("get vm error %v", err)
	}

	if update {
		_, _, err = client.ApplyVMI(vmi, false)
		if err != nil {
			return fmt.Errorf("update vm %s dataVolume error %v", vm.GetName(), err)
		}
	} else {
		_, err = client.KubevirtCli.
			VirtualMachineInstance(vm.GetWorkspace()).
			Create(vmi)

		if err != nil {
			return fmt.Errorf("create vm %s dataVolume error %v", vm.GetName(), err)
		}
	}
	return nil
}

func (V *VMCtrl) changeVMStatus(vm *compute.VirtualMachine, status string, message string) error {
	vm.Spec.Status = status
	vm.Spec.Message = message
	_, _, err := V.stage.Apply(common.DefaultDatabase, common.VIRTUALMACHINE, vm.GetName(), vm, false)
	return err
}

func (V *VMCtrl) updateVMIToK8s(vm *compute.VirtualMachine) {
	flog := V.flog.WithField("func", "updateVMIToK8s")

	client, err := V.cs.GetClient(common.DefaultKubernetes)

	_, err = client.KubevirtCli.VirtualMachineInstance(vm.GetWorkspace()).Get(vm.GetName(), &metav1.GetOptions{})
	if err != nil {
		flog.Warnf("get vmi %s error %v", vm.GetName(), err)
		return
	}

	switch vm.Spec.State {
	case compute.Updating:
		if err := V.CreateOrApplyVMI(client, vm); err != nil {
			flog.Warnf("update vmi %s error %v", vm.GetName(), err)
		}

	case compute.Stopping:
		err = client.KubevirtCli.VirtualMachineInstance(vm.GetWorkspace()).Pause(vm.GetName())
		if err != nil {
			flog.Warnf("stop vmi %s error %v", vm.GetName(), err)
			if applyErr := V.changeVMStatus(vm, common.FAIL, err.Error()); applyErr != nil {
				flog.Warnf("change vm status error %v", applyErr)
			}
		}

	case compute.Starting:
		err = client.KubevirtCli.VirtualMachineInstance(vm.GetWorkspace()).Unpause(vm.GetName())
		if err != nil {
			flog.Warnf("start vmi %s error %v", vm.GetName(), err)
			if applyErr := V.changeVMStatus(vm, common.FAIL, err.Error()); applyErr != nil {
				flog.Warnf("change vm status error %v", applyErr)
			}
		}
	}
}

func (V *VMCtrl) checkObjStatusAndGetObj(virtualMachine *unstructured.Unstructured, checkStatus ...string) (bool, *compute.VirtualMachine) {
	status := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "status")

	for _, s := range checkStatus {
		if status == s {
			return false, nil
		}
	}

	id := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "instanceId")
	vpcId := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "vpcId")
	vSwitchId := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "vSwitchId")
	az := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "az")
	region := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "regionId")
	instanceType := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "instanceType")
	image := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "image")
	state := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "state")
	message := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "message")
	localName := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "localName")
	cpu := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "cpu")
	rootDevice := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "rootDevice")
	memory := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "memory")
	createTime := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "createTime")
	os := objUtils.GetNestedString(virtualMachine.Object,
		"spec", "os")

	spec := objUtils.GetNestedMap(virtualMachine.Object,
		"spec")

	labels := map[string]interface{}{}
	getLabels := virtualMachine.GetLabels()
	for k, v := range getLabels {
		labels[k] = v
	}

	vm := &compute.VirtualMachine{
		Metadata: core.Metadata{
			Name:      virtualMachine.GetName(),
			Kind:      core.Kind(virtualMachine.GetKind()),
			Namespace: virtualMachine.GetNamespace(),
			Workspace: getLabels["workspace"],
			Labels:    labels,
		},
		Spec: compute.VirtualMachineSpec{
			CPU:          cpu,
			Memory:       memory,
			Vendor:       virtualMachine.GetNamespace(),
			InstanceId:   id,
			RootDevice:   rootDevice,
			Az:           az,
			RegionId:     region,
			VPCId:        vpcId,
			VSwitchId:    vSwitchId,
			InstanceType: instanceType,
			ImageId:      image,
			Status:       status,
			State:        compute.VirtualMachineStateType(state),
			Message:      message,
			LocalName:    localName,
			Os:           os,
			CreateTime:   createTime,
		},
	}

	securityGroup := spec["securityGroup"]
	if securityGroup != nil {
		_securityGroups := make([]string, 0)
		for _, _securityGroup := range securityGroup.([]interface{}) {
			_securityGroups = append(_securityGroups, _securityGroup.(string))
		}
		vm.Spec.SecurityGroup = _securityGroups
	}

	vmNetworkInterfaces := make([]networking.NetworkInterface, 0)
	filter := map[string]interface{}{
		"spec.attachment.instance_id": vm.Spec.InstanceId,
	}
	err := V.stage.ListToObject(common.DefaultDatabase, common.NETWORKINTERFACE, filter, &vmNetworkInterfaces, true)
	if err != nil {
		return false, nil
	}
	for _, _networkInterface := range vmNetworkInterfaces {
		networkInterface := compute.NetWorkInterface{
			NetworkInterfaceId: _networkInterface.Spec.ID,
			VpcId:              _networkInterface.Spec.VPCId,
			VSwitchId:          _networkInterface.Spec.SubnetId,
			Mac:                _networkInterface.Spec.MacAddress,
			Type:               _networkInterface.Spec.Type,
			SecureGroups:       _networkInterface.Spec.SecurityGroupIds,
			Status:             _networkInterface.Spec.Status,
			State:              _networkInterface.Spec.State,
			PrivateIp:          _networkInterface.Spec.PrivateIpAddress,
			PublicIp:           _networkInterface.Spec.PublicIpAddress,
			Message:            _networkInterface.Spec.Message,
		}
		for _, set := range _networkInterface.Spec.PrivateIpSets {
			if !set.Primary {
				networkInterface.SecondaryPrivateIp = append(networkInterface.SecondaryPrivateIp, set.PrivateIpAddress)
			}
		}
		vm.Spec.NetWorkInterface = append(vm.Spec.NetWorkInterface, networkInterface)
		vm.Spec.PrivateIpAddress = append(vm.Spec.PrivateIpAddress, networkInterface.PrivateIp)
		vm.Spec.PublicIpAddress = append(vm.Spec.PublicIpAddress, networkInterface.PublicIp)
	}

	vmStorages := make([]compute.Storage, 0)
	filter = map[string]interface{}{
		"spec.attachments.instance_id": vm.Spec.InstanceId,
	}
	if err := V.stage.ListToObject(common.DefaultDatabase, common.STORAGE, filter, &vmStorages, true); err != nil {
		return false, nil
	}
	for _, _storage := range vmStorages {
		vmStorage := compute.VmStorage{
			DeleteOnTermination: _storage.Spec.DeleteWithInstance,
			Iops:                int64(_storage.Spec.IOPS),
			Throughput:          int64(_storage.Spec.Throughput),
			VolumeId:            _storage.Spec.StorageId,
			VolumeSize:          int64(_storage.Spec.Size),
			VolumeType:          _storage.Spec.CategoryType,
			DiskType:            _storage.Spec.DiskType,
			State:               _storage.Spec.State,
			Status:              _storage.Spec.Status,
			Message:             _storage.Spec.Message,
		}
		for _, attachments := range _storage.Spec.Attachments {
			if attachments.InstanceId == vm.Spec.InstanceId {
				if vm.GetNamespace() == common.AWS {
					if attachments.Device == vm.Spec.RootDevice {
						vmStorage.DiskType = "system"
					} else {
						vmStorage.DiskType = "data"
					}
				}
				vmStorage.State = attachments.State
			}
		}
		vm.Spec.Storage = append(vm.Spec.Storage, vmStorage)
	}

	if vm.Workspace == "" {
		vm.Workspace = common.DefaultWorkspace
	}

	return true, vm
}
