package dep

import (
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func int32Ptr(i int32) *int32 { return &i }

// CreateDep function to get pods
func CreateDep(clientset kubernetes.Interface) {
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	result, err := deploymentsClient.Get(deployment.Name, metav1.GetOptions{})
	if err != nil || result.GetName() == "" {
		result, err := deploymentsClient.Create(deployment)
		fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
		if err != nil {
			fmt.Println(err)
		}
	}
}
func int64Ptr(i int64) *int64 { return &i }

// WatchDep function to get pods
func WatchDep(clientset *kubernetes.Clientset) {
	depClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	watcher, err := depClient.Watch(
		metav1.ListOptions{
			TimeoutSeconds: int64Ptr(10),
		})
	if err != nil {
		fmt.Println(err)
	}
	ch := watcher.ResultChan()
	for event := range ch {
		svc, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			watcher.Stop()
			log.Fatal("unexpected type")
		}
		log.Println(svc)
		switch event.Type {
		case watch.Added:
			fmt.Println("Deployment added\n", svc.GetName())
		case watch.Modified:
			fmt.Println("Deployment modified\n", svc.GetName())
		case watch.Deleted:
			fmt.Println("Deployment deleted\n", svc.GetName())
		case watch.Error:
			fmt.Println("watcher error encountered\n", svc.GetName())
		}
	}
}
