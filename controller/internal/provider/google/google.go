package google

import (
	"context"
	"fmt"
	"net/http"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	mcDns "github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type GoogleDnsClient struct {
	config provider.Config
	*dns.Service
}

var (
	DefaultTtl      int64 = 60 * 30
	DefaultPriority int   = 0
	DefaultWeight   int   = 0
)

func (c *GoogleDnsClient) SetExternalDns(hostname string, gs *agonesv1.GameServer) (provider.ServerResponse, error) {
	change := dns.Change{}

	aRecord := NewARecordSet(hostname, gs, DefaultTtl)
	srvRecord, err := NewSrvRecordSet(hostname, gs, DefaultTtl)
	if err != nil {
		return provider.ServerResponse{}, err
	}

	change.Additions = []*dns.ResourceRecordSet{aRecord, srvRecord}
	res, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	if err != nil {
		apiError, ok := err.(*googleapi.Error)
		if ok {
			res := provider.ServerResponse{HTTPStatusCode: apiError.Code, Header: apiError.Header}

			if apiError.Code == http.StatusConflict || apiError.Code == http.StatusNotModified {
				return res, &provider.DNSRecordExists{Record: aRecord.Name}
			}

			return res, apiError
		}

		return provider.ServerResponse{}, err
	}

	serverRes := provider.ServerResponse{HTTPStatusCode: res.HTTPStatusCode, Header: res.Header}

	return serverRes, nil
}

func (c *GoogleDnsClient) RemoveExternalDns(hostname string, gs *agonesv1.GameServer) (provider.ServerResponse, error) {
	change := dns.Change{}

	aRecord := NewARecordSet(hostname, gs, DefaultTtl)
	srvRecord, err := NewSrvRecordSet(hostname, gs, DefaultTtl)
	if err != nil {
		return provider.ServerResponse{}, err
	}

	change.Deletions = []*dns.ResourceRecordSet{srvRecord, aRecord}
	res, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	if err != nil {
		apiError, ok := err.(*googleapi.Error)
		if ok {
			return provider.ServerResponse{HTTPStatusCode: apiError.Code, Header: apiError.Header},
				&provider.DNSRecordNonExistent{Records: []string{aRecord.Name, srvRecord.Name}, ServerError: err}
		}

		return provider.ServerResponse{}, err
	}

	return provider.ServerResponse{HTTPStatusCode: res.HTTPStatusCode, Header: res.Header}, nil
}

func NewSrvRecordSet(hostname string, gs *agonesv1.GameServer, ttl int64) (*dns.ResourceRecordSet, error) {
	ports := gs.Status.Ports

	if ports == nil {
		return nil, fmt.Errorf("server ports not allocated")
	}

	port := gs.Status.Ports[0].Port
	recordName := mcDns.JoinSrvRecordName(hostname, gs.Name)
	resourceRecord := mcDns.JoinSrvRR(recordName, uint16(port), DefaultPriority, DefaultWeight)
	return &dns.ResourceRecordSet{Type: "SRV", Name: recordName, Rrdatas: []string{resourceRecord}, Ttl: ttl}, nil
}

func NewARecordSet(hostname string, gs *agonesv1.GameServer, ttl int64) *dns.ResourceRecordSet {
	recordName := mcDns.JoinARecordName(hostname, gs.Name)
	hostExternalIp := gs.Status.Address

	return &dns.ResourceRecordSet{Type: "A", Name: recordName, Rrdatas: []string{hostExternalIp}, Ttl: ttl}
}

func NewDnsClient(projectId, managedZone string) (*GoogleDnsClient, error) {
	gcloud, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, err
	}

	dns, err := dns.NewService(context.Background(), option.WithHTTPClient(gcloud))

	if err != nil {
		return nil, err
	}

	config := provider.Config{GoogleProjectId: projectId, GoogleManagedZone: managedZone}

	return &GoogleDnsClient{config: config, Service: dns}, nil
}
