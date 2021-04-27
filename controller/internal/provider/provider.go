package provider

import (
	"net/http"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	GoogleProjectId   string
	GoogleManagedZone string
}

type ServerResponse struct {
	HTTPStatusCode int
	Header         http.Header
}

type DnsClient interface {
	SetGameServerExternalDns(hostname string, gs *agonesv1.GameServer) error
	RemoveGameServerExternalDns(hostname string, gs *agonesv1.GameServer) error
	SetNodeExternalDns(hostname string, node *corev1.Node) error
	RemoveNodeExternalDns(hostname string, node *corev1.Node) error
	IgnoreClientError(err error) error
	IgnoreAlreadyExists(err error) error
}
