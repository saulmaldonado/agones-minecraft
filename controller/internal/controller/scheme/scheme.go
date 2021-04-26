package scheme

import (
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	corev1 "k8s.io/api/core/v1"
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

func GetNodeExternalAddress(node *corev1.Node) (string, error) {
	addresses := node.Status.Addresses

	for _, address := range addresses {
		if address.Type == corev1.NodeExternalIP {
			return address.Address, nil
		}
	}

	return "", &NoNodeExternalIP{node.Name}
}

func GetNodeExternalDNS(node *corev1.Node) (string, bool) {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeExternalDNS {
			return address.Address, true
		}
	}

	return "", false
}

type NoNodeExternalIP struct {
	NodeName string
}

func (e *NoNodeExternalIP) Error() string {
	return fmt.Sprintf("%s has no external IP", e.NodeName)
}
