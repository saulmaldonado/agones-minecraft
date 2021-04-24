package google

import (
	"context"
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
	port := gs.Status.Ports[0].Port

	aRecordName := mcDns.JoinARecordName(hostname, gs.Name)
	srvRecordName := mcDns.JoinSrvRecordName(hostname, gs.Name)

	srvResourceRecord := mcDns.JoinSrvRR(srvRecordName, uint16(port), DefaultPriority, DefaultWeight)

	srvRecord := NewSrvRecordSet(srvRecordName, srvResourceRecord, DefaultTtl)
	aRecord := NewARecordSet(aRecordName, gs.Status.Address, DefaultTtl)

	change.Additions = []*dns.ResourceRecordSet{aRecord, srvRecord}
	res, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	if err != nil {
		apiError, ok := err.(*googleapi.Error)
		if ok {
			res := provider.ServerResponse{HTTPStatusCode: apiError.Code, Header: apiError.Header}

			if apiError.Code == http.StatusConflict || apiError.Code == http.StatusNotModified {
				return res, &provider.DNSRecordExists{Record: aRecordName}
			}

			return res, apiError
		}

		return provider.ServerResponse{}, err
	}

	serverRes := provider.ServerResponse{HTTPStatusCode: res.HTTPStatusCode, Header: res.Header}

	return serverRes, nil
}

func NewSrvRecordSet(srvRecordName string, srvResourceRecord string, ttl int64) *dns.ResourceRecordSet {
	return &dns.ResourceRecordSet{Type: "SRV", Name: srvRecordName, Rrdatas: []string{srvResourceRecord}, Ttl: ttl}
}

func NewARecordSet(recordName string, hostExternalIp string, ttl int64) *dns.ResourceRecordSet {
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
