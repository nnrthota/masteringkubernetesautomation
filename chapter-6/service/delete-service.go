package service

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeleteService function to get pods
func DeleteService(clientset *kubernetes.Clientset, name string) {
	servicesClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	fmt.Println("Deleting service...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := servicesClient.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Deleted service.")
}
