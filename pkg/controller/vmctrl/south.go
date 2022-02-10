package vmctrl

import (
	"context"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	objUtils "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func (V *VMCtrl) SouthOnAdd(obj runtime.Object) {
	groupKind := obj.GetObjectKind().GroupVersionKind().GroupKind()

	switch groupKind.Group {
	case "github.com/ddx2x":
		V.addThirdVmToStage(obj)
	case "kubevirt.io":
		V.applyLiZiVmToStage(obj)
	}
}

func (V *VMCtrl) SouthOnUpdate(obj runtime.Object) {
	groupKind := obj.GetObjectKind().GroupVersionKind().GroupKind()

	switch groupKind.Group {
	case "github.com/ddx2x":
		V.updateThirdVmToStage(obj)
	case "kubevirt.io":
		V.applyLiZiVmToStage(obj)
	}
}

func (V *VMCtrl) SouthOnDelete(obj runtime.Object) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	flog := log.G(ctx).WithField("controller", "SouthOnDelete")
	virtualMachine := &unstructured.Unstructured{}

	err := objUtils.Unmarshal(virtualMachine, obj)
	if err != nil {
		flog.Warnf("unmarshal virtualMachine data error: %v", obj)
	}

	labels := virtualMachine.GetLabels()

	err = V.stage.Delete(common.DefaultDatabase, common.VIRTUALMACHINE, virtualMachine.GetName(), labels["workspace"])
	if err != nil {
		flog.Warnf("delete virtualMachine error: %v", obj)
	}
	flog.Infof("delete a virtualMachine %s from stage", virtualMachine.GetName())
}

func (V *VMCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
	flog := V.flog.WithField("func", "SouthEventChs")
	flog.Infof("vm start watch south event")

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

		watchInterface, err := client.Interface.Resource(virtualMachineGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())

		// 增加对 kubeVirt 的 watch
		virtWatchInterface, err := client.Interface.Resource(virtualMachineInstanceGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, virtWatchInterface.ResultChan())
	}

	return channels, nil
}
