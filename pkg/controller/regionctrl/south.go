package regionctrl

import (
	"context"
	"fmt"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func (V *RegionCtrl) SouthOnAdd(obj runtime.Object) {
	V.applyRegionToStage(obj)

}

func (V *RegionCtrl) SouthOnUpdate(obj runtime.Object) {
	V.applyRegionToStage(obj)

}

func (V *RegionCtrl) SouthOnDelete(obj runtime.Object) {
	panic("implement me")
}

func (V *RegionCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
	flog := V.flog.WithField("func", "SouthEventChs")
	flog.Info("region start watch South Event")

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

		watchInterface, err := client.Interface.Resource(regionGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}

func (V *RegionCtrl) applyRegionToStage(obj runtime.Object) {
	flog := V.flog.WithField("func", "applyRegionToStage")

	region := &unstructured.Unstructured{}
	err := utilsObj.Unmarshal(region, obj)
	if err != nil {
		flog.Warnf("unmarshal region data error: %v", obj)
		return
	}

	localName := utilsObj.GetNestedString(region.Object,
		"spec", "LocalName")

	endPoint := utilsObj.GetNestedString(region.Object,
		"spec", "regionEndpoint")

	id := utilsObj.GetNestedString(region.Object,
		"spec", "regionId")

	regionObj := &system.Region{
		Metadata: core.Metadata{
			Name:      fmt.Sprintf("%s-%s", region.GetNamespace(), region.GetName()),
			Namespace: region.GetNamespace(),
			Workspace: common.DefaultWorkspace,
			Kind:      core.Kind(region.GetKind()),
		},
		Spec: system.RegionSpec{
			LocalName: localName,
			Endpoint:  endPoint,
			ID:        id,
		},
	}

	err = V.stage.Get(common.DefaultDatabase, common.REGION, regionObj.Name, &system.Region{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.REGION, regionObj)
		if createErr != nil {
			flog.Warnf("create region error %v", createErr)
			return
		}
		flog.Infof("create region %s, namespace %s", region.GetName(), region.GetNamespace())
		return
	}
	if err != nil {
		flog.Warnf("get region error %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.REGION, regionObj.Name, regionObj, false)
	if err != nil {
		flog.Warnf("update region error: %v", err)
		return
	}

	if update {
		flog.Infof("update region namespace: %s, name: %s", regionObj.GetNamespace(), regionObj.GetName())
	}
	return
}
