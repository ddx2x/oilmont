package zonectrl

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

func (V *ZoneCtrl) SouthOnAdd(obj runtime.Object) {
	V.applyZoneToStage(obj)

}

func (V *ZoneCtrl) SouthOnUpdate(obj runtime.Object) {
	V.applyZoneToStage(obj)
}

func (V *ZoneCtrl) SouthOnDelete(obj runtime.Object) {
	panic("implement me")
}

func (V *ZoneCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
	flog := V.flog.WithField("func", "SouthEventChs")
	flog.Info("zone start watch South Event")

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

		watchInterface, err := client.Interface.Resource(zoneGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}

func (V *ZoneCtrl) applyZoneToStage(obj runtime.Object) {
	flog := V.flog.WithField("func", "applyZoneToStage")

	zone := &unstructured.Unstructured{}
	err := utilsObj.Unmarshal(zone, obj)
	if err != nil {
		flog.Warnf("unmarshal zone data error: %v", obj)
		return
	}

	region := utilsObj.GetNestedString(zone.Object,
		"metadata", "labels", "region")

	localName := utilsObj.GetNestedString(zone.Object,
		"spec", "localName")

	id := utilsObj.GetNestedString(zone.Object,
		"spec", "zoneId")

	zoneObj := &system.AvailableZone{
		Metadata: core.Metadata{
			Name:      fmt.Sprintf("%s-%s", zone.GetNamespace(), zone.GetName()),
			Namespace: zone.GetNamespace(),
			Workspace: common.DefaultWorkspace,
			Kind:      core.Kind(zone.GetKind()),
		},
		Spec: system.AvailableZoneSpec{
			Region:    region,
			LocalName: localName,
			ID:        id,
		},
	}

	err = V.stage.Get(common.DefaultDatabase, common.AVAILABLEZONE, zoneObj.GetName(), &system.AvailableZone{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.AVAILABLEZONE, zoneObj)
		if createErr != nil {
			flog.Warnf("create zone error: %v", createErr)
			return
		}
		flog.Infof("create zone %s, namespace %s", zoneObj.GetName(), zoneObj.GetNamespace())
		return
	}
	if err != nil {
		flog.Warnf("get zone error: %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.AVAILABLEZONE, zoneObj.GetName(), zoneObj, false)
	if err != nil {
		flog.Warnf("update zone error: %v", err)
		return
	}
	if update {
		flog.Infof("update new zone %s, namespace %s", zoneObj.GetName(), zoneObj.GetNamespace())
	}
}
