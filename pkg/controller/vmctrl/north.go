package vmctrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/resource/compute"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
)

func (V *VMCtrl) NorthOnAdd(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnAdd")

	vm := &compute.VirtualMachine{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, vm); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}
	if vm.Spec.Status != common.INIT {
		return
	}

	switch vm.GetNamespace() {
	case common.AWS, common.ALIYUN:
		V.addThirdPartyToK8s(vm)
	default:
		V.applyLiZiVmToK8s(vm)
	}

}

func (V *VMCtrl) NorthOnUpdate(obj core.IObject) {
	flog := V.flog.WithField("func", "NorthOnUpdate")

	vm := &compute.VirtualMachine{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, vm); err != nil {
		flog.Warnf("unstructured obj error %v", err)
		return
	}
	if vm.Spec.Status != common.UPDATE {
		return
	}

	switch vm.Spec.Vendor {
	case common.AWS, common.ALIYUN:
		V.updateThirdPartyToK8s(vm)
	default:
		V.updateVMIToK8s(vm)
	}
}

func (V *VMCtrl) NorthOnDelete(obj core.IObject) {
	vm := &compute.VirtualMachine{}
	if err := utilsObj.UnstructuredObjectToInstanceObj(obj, vm); err != nil {
		return
	}

	switch vm.GetNamespace() {
	case common.AWS, common.ALIYUN:
		V.deleteK8sDataThirdPartyVM(vm)

	default:
		V.deleteLiZiVM(vm)
	}
}

func (V *VMCtrl) NorthEventCh(ctx context.Context) (<-chan core.Event, error) {
	return V.stage.WatchEvent(ctx, common.DefaultDatabase, common.VIRTUALMACHINE, "0")
}
