package provider

import (
	"fmt"
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

type DNSRecordExists struct {
	Record string
}

func (e *DNSRecordExists) Error() string {
	return fmt.Sprintf("record for %s already exists", e.Record)
}

type DNSRecordNonExistent struct {
	Records     []string
	ServerError error
}

func (e *DNSRecordNonExistent) Error() string {
	return fmt.Sprintf("records for %v non existent", e.Records)
}
