package controller

import (
	"github.com/example-inc/sample-operator/pkg/controller/sampleop"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, sampleop.Add)
}
