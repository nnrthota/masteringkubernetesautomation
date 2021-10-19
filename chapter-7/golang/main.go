package main

import (
	"fmt"

	"github.com/narendranathreddythota/masteringkubernetesautomation/chapter7/dep"
	"github.com/narendranathreddythota/masteringkubernetesautomation/chapter7/helper"
	"github.com/narendranathreddythota/masteringkubernetesautomation/chapter7/pvc"
	"github.com/narendranathreddythota/masteringkubernetesautomation/chapter7/secret"
	svc "github.com/narendranathreddythota/masteringkubernetesautomation/chapter7/service"
)

func main() {

	// create the clientset
	clientset, err := helper.Client()
	if err != nil {
		fmt.Printf(err.Error())
	}
	dep.CreateDep(clientset, dep.CreateWordpressDep())
	dep.CreateDep(clientset, dep.CreateMySQLDep())
	secret.CreateSecret(clientset, secret.CreateWordpressSecret())
	pvc.CreatePvc(clientset, pvc.CreateWordpressPVC())
	pvc.CreatePvc(clientset, pvc.CreateMYSQLPVC())
	svc.CreateService(clientset, svc.CreateWordpressService())
	svc.CreateService(clientset, svc.CreateMYSQLService())
}
