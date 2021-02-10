package controller

import (
	"celery-operator/pkg/controller/celery"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, celery.Add)
}
