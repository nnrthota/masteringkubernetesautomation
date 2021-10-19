package pvc

import (
	"fmt"
	"log"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// CreatePvc function to get pods
func CreatePvc(clientset *kubernetes.Clientset) {
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault)
	class := "manual"
	pvc := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-pvc",
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes:      []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
			StorageClassName: &class,
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceName(apiv1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}
	result, err := pvcClient.Get(pvc.Name, metav1.GetOptions{})
	if err != nil || result.GetName() == "" {
		result, err := pvcClient.Create(pvc)
		fmt.Printf("Created PVC %q.\n", result.GetObjectMeta().GetName())
		if err != nil {
			fmt.Println(err)
		}
	}

}
func int64Ptr(i int64) *int64 { return &i }

var totalClaimedSize resource.Quantity

// WatchPVC function to get pods
func WatchPVC(clientset *kubernetes.Clientset) {
	maxClaims := "30Gi"
	maxClaimedSize := resource.MustParse(maxClaims)
	depClient := clientset.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault)
	watcher, err := depClient.Watch(
		metav1.ListOptions{
			TimeoutSeconds: int64Ptr(10),
		})
	if err != nil {
		fmt.Println(err)
	}
	ch := watcher.ResultChan()
	for event := range ch {
		pvc, ok := event.Object.(*apiv1.PersistentVolumeClaim)
		if !ok {
			fmt.Println("unexpected type assertion")
		}
		fmt.Println(pvc)
		size := pvc.Spec.Resources.Requests[v1.ResourceStorage]
		switch event.Type {
		case watch.Added:
			totalClaimedSize.Add(size)
			fmt.Println("PVC added\n", pvc.GetName())
			if totalClaimedSize.Cmp(maxClaimedSize) == 1 {
				log.Printf("\nClaim quota reached: max %s at %s",
					maxClaimedSize.String(),
					totalClaimedSize.String(),
				)
			}
		case watch.Modified:
			fmt.Println("PVC modified\n", pvc.GetName())
		case watch.Deleted:
			totalClaimedSize.Sub(size)
			fmt.Println("PVC deleted\n", pvc.GetName())
		case watch.Error:
			fmt.Println("watcher error encountered\n", pvc.GetName())
		}
		log.Printf("\n At %3.1f%% claim capacity (%s/%s)\n",
			float64(totalClaimedSize.Value())/float64(maxClaimedSize.Value())*100,
			totalClaimedSize.String(),
			maxClaimedSize.String(),
		)
	}
}
