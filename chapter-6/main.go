package main

import (
	"fmt"

	"github.com/narendranath/kube-go/dep"
	"github.com/narendranath/kube-go/helper"
	"github.com/narendranath/kube-go/pvc"
	"github.com/narendranath/kube-go/service"
)

func main() {

	// create the clientset
	clientset, err := helper.Client()
	if err != nil {
		fmt.Printf(err.Error())
	}

	dep.CreateDep(clientset)
	dep.WatchDep(clientset)
	service.CreateService(clientset)
	service.WatchService(clientset)
	pvc.CreatePvc(clientset)
	pvc.WatchPVC(clientset)
}
