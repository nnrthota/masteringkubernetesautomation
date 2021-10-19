package pvc

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreatePvc function to get pods
func CreatePvc(clientset *kubernetes.Clientset, pvc apiv1.PersistentVolumeClaim) {
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault)
	// Create Service
	fmt.Println("Creating pvc...")
	result, err := pvcClient.Create(&pvc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created PVC %q.\n", result.GetObjectMeta().GetName())
}

func CreateWordpressPVC() apiv1.PersistentVolumeClaim {
	pvc := apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "wp-pv-claim",
			Labels: map[string]string{"app": "wordpress"},
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceName(apiv1.ResourceStorage): resource.MustParse("4Gi"),
				},
			},
		},
	}
	return pvc
}

func CreateMYSQLPVC() apiv1.PersistentVolumeClaim {
	pvc := apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "mysql-pv-claim",
			Labels: map[string]string{"app": "wordpress"},
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceName(apiv1.ResourceStorage): resource.MustParse("4Gi"),
				},
			},
		},
	}
	return pvc
}

//func int32Ptr(i int32) *int32 { return &i }
