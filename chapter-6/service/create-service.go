package service

import (
	"fmt"
	"log"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// CreateService function to get pods
func CreateService(clientset *kubernetes.Clientset) {
	servicesClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-service",
		},
		Spec: apiv1.ServiceSpec{
			Type: "NodePort",
			Selector: map[string]string{
				"app": "demo",
			},
			Ports: []apiv1.ServicePort{
				{
					Name:     "web",
					Protocol: apiv1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 80,
					},
					NodePort: 30080,
				},
			},
		},
	}
	result, err := servicesClient.Get(service.Name, metav1.GetOptions{})
	if err != nil || result.GetName() == "" {
		service, err := servicesClient.Create(service)
		fmt.Printf("Created Service %q.\n", service.GetObjectMeta().GetName())
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("result.Name", result.Name)
}
func int64Ptr(i int64) *int64 { return &i }

// WatchService function to get pods
func WatchService(clientset *kubernetes.Clientset) {
	servicesClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	watcher, err := servicesClient.Watch(
		metav1.ListOptions{
			TimeoutSeconds: int64Ptr(10),
		})
	if err != nil {
		fmt.Println(err)
	}
	ch := watcher.ResultChan()
	for event := range ch {
		svc, ok := event.Object.(*v1.Service)
		if !ok {
			watcher.Stop()
			log.Fatal("unexpected type")
		}
		log.Println(svc)
		switch event.Type {
		case watch.Added:
			fmt.Println("Service added\n", svc.GetName())
		case watch.Modified:
			fmt.Println("Service modified\n", svc.GetName())
		case watch.Deleted:
			fmt.Println("Service deleted\n", svc.GetName())
		case watch.Error:
			fmt.Println("watcher error encountered\n", svc.GetName())
		}
	}
}
