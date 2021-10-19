package pvc

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeletePvc function to get pods
func DeletePvc(clientset *kubernetes.Clientset, name string) {
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault)
	fmt.Println("Deleting pvc...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := pvcClient.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Deleted pvc.")
}
