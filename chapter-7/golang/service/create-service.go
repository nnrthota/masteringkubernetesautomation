package service

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// CreateService function to get pods
func CreateService(clientset *kubernetes.Clientset, service apiv1.Service) {
	servicesClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	// Create Service
	fmt.Println("Creating service...")

	result, err := servicesClient.Get(service.Name, metav1.GetOptions{})
	if err != nil || result.GetName() == "" {
		service, err := servicesClient.Create(&service)
		fmt.Printf("Created Service %q.\n", service.GetObjectMeta().GetName())
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("result.Name", result.Name)
}

func CreateWordpressService() apiv1.Service {
	service := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wordpress",
		},
		Spec: apiv1.ServiceSpec{
			Type: "NodePort",
			Selector: map[string]string{
				"app":  "wordpress",
				"tier": "frontend",
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
	return service
}

func CreateMYSQLService() apiv1.Service {
	service := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "wordpress-mysql",
			Labels: map[string]string{"app": "wordpress"},
		},
		Spec: apiv1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app":  "wordpress",
				"tier": "mysql",
			},
			Ports: []apiv1.ServicePort{
				{
					Name:     "msql",
					Protocol: apiv1.ProtocolTCP,
					Port:     3306,
				},
			},
		},
	}
	return service
}

//func int32Ptr(i int32) *int32 { return &i }
