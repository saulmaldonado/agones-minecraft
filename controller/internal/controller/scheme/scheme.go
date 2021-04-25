package scheme

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
)

func AddToScheme(scheme *runtime.Scheme) error {
	if err := agonesv1.AddToScheme(scheme); err != nil {
		return err
	}

	if err := clientGoScheme.AddToScheme(scheme); err != nil {
		return err
	}

	return nil
}
