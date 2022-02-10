package main

import (
	"context"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"
)

func main() {
	ctx := context.TODO()

	k8s.ShardingResourceRegistry.Register("pods", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"})

	cfg, err := k8s.NewKubeConfigure(ctx, k8s.ShardingResourceRegistry, *common.KubeConfig, common.InCluster)
	if err != nil {
		panic(err)
	}

	_ = cfg

}

func dvCreate() {
	quantity, err := resource.ParseQuantity("6Gi")
	if err != nil {
		panic(err)
	}
	dv := &v1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataVolume",
			APIVersion: "cdi.kubevirt.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test3",
			Namespace: "default",
		},
		Spec: v1beta1.DataVolumeSpec{
			Source: &v1beta1.DataVolumeSource{
				Registry: &v1beta1.DataVolumeSourceRegistry{
					URL: "docker://laiks/fedora:cloud-base",
				},
			},

			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: quantity,
					},
				},
			},
		},
	}
	_ = dv
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//new, err := virtClient.CdiClient().CdiV1beta1().DataVolumes("default").Create(ctx, dv, metav1.CreateOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//_ = new

}
