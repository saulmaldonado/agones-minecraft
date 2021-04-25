package provider

import (
	"fmt"
	"net/http"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
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
	SetExternalDns(hostname string, gs *agonesv1.GameServer) (ServerResponse, error)
	RemoveExternalDns(hostname string, gs *agonesv1.GameServer) (ServerResponse, error)
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
