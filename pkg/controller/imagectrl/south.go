package imagectrl

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/compute"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	utilsObj "github.com/ddx2x/oilmont/pkg/utils/obj"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func (V *ImageCtrl) SouthOnAdd(obj runtime.Object) {
	V.applyImageToStage(obj)

}

func (V *ImageCtrl) SouthOnUpdate(obj runtime.Object) {
	V.applyImageToStage(obj)

}

func (V *ImageCtrl) SouthOnDelete(obj runtime.Object) {
	V.deleteImageToStage(obj)
}

func (V *ImageCtrl) SouthEventChs(ctx context.Context) ([]<-chan watch.Event, error) {
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

		watchInterface, err := client.Interface.Resource(imageGvr).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		channels = append(channels, watchInterface.ResultChan())
	}

	return channels, nil
}

func (V *ImageCtrl) applyImageToStage(obj runtime.Object) {
	var os string
	flog := V.flog.WithField("func", "applyImageToStage")

	image := &unstructured.Unstructured{}
	err := utilsObj.Unmarshal(image, obj)
	if err != nil {
		flog.Warnf("unmarshal image data error: %v", obj)
		return
	}

	os = utilsObj.GetNestedString(image.Object,
		"spec", "os")

	region := utilsObj.GetNestedString(image.Object,
		"spec", "region")

	id := utilsObj.GetNestedString(image.Object,
		"spec", "id")

	if image.GetNamespace() == common.AWS {
		os = image.GetName()
	}

	imageObj := &compute.Image{
		Metadata: core.Metadata{
			Name:      image.GetName(),
			Namespace: image.GetNamespace(),
			Workspace: common.DefaultWorkspace,
		},
		Spec: compute.ImageSpec{
			Os:     os,
			Region: region,
			ID:     id,
		},
	}

	err = V.stage.Get(common.DefaultDatabase, common.IMAGE, imageObj.Name, &compute.Image{}, false)
	if err == datasource.NotFound {
		_, createErr := V.stage.Create(common.DefaultDatabase, common.IMAGE, imageObj)
		if createErr != nil {
			flog.Warnf("create image error %v", createErr)
			return
		}
		flog.Infof("create image %s, namespace %s", imageObj.GetName(), imageObj.GetNamespace())
		return
	}
	if err != nil {
		flog.Warnf("get image error %v", err)
		return
	}

	_, update, err := V.stage.Apply(common.DefaultDatabase, common.IMAGE, imageObj.Name, imageObj, false)
	if err != nil {
		flog.Warnf("update image error: %v", err)
		return
	}
	if update {
		flog.Infof("update image %s, namespace: %s", imageObj.GetName(), imageObj.GetNamespace())
	}
	return
}

func (V *ImageCtrl) deleteImageToStage(obj runtime.Object) {
	flog := V.flog.WithField("func", "deleteImageToStage")
	image := &unstructured.Unstructured{}

	err := utilsObj.Unmarshal(image, obj)
	if err != nil {
		flog.Warnf("unmarshal image data error: %v", obj)
	}

	err = V.stage.Delete(common.DefaultDatabase, common.IMAGE, image.GetName(), common.DefaultWorkspace)
	if err != nil {
		flog.Warnf("delete image error: %v", obj)
	}
	flog.Infof("delete a image %s from stage", image.GetName())
}
