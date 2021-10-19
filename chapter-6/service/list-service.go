package service

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListService function to get services
func ListService(clientset *kubernetes.Clientset) {
	servicesClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	fmt.Printf("Listing services in namespace %q:\n", apiv1.NamespaceDefault)
	list, err := servicesClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s \n", d.Name)
	}
}
