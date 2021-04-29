package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Finalizer = "kubernetes.io/agones-mc"

func findFinalizer(obj client.Object) bool {
	finalizers := obj.GetFinalizers()
	for _, f := range finalizers {
		if f == Finalizer {
			return true
		}
	}
	return false
}

func setFinalizer(obj client.Object) {
	finalizers := obj.GetFinalizers()

	if finalizers == nil {
		finalizers = []string{}
	}

	obj.SetFinalizers(append(finalizers, Finalizer))
}

func removeFinalizer(obj client.Object) {
	finalizers := obj.GetFinalizers()
	newFinalizers := []string{}
	for _, f := range finalizers {
		if f != Finalizer {
			newFinalizers = append(newFinalizers, f)
		}
	}
	obj.SetFinalizers(newFinalizers)
}
