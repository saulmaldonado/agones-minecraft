package google

import (
	"context"
	"net/http"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/controller/scheme"
	mcDns "github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
)

type GoogleDnsClient struct {
	config provider.Config
	*dns.Service
}

const (
	DefaultTtl      int64  = 60 * 30
	DefaultPriority int    = 0
	DefaultWeight   int    = 0
	SRV             string = "SRV"
	A               string = "A"
)

func (c *GoogleDnsClient) SetExternalDns(hostname string, gs *agonesv1.GameServer) (provider.ServerResponse, error) {
	change := dns.Change{}

	srvRecord := NewSrvRecordSet(hostname, gs, DefaultTtl)

	change.Additions = []*dns.ResourceRecordSet{srvRecord}
	res, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	if err != nil {

		switch e := err.(type) {
		case *googleapi.Error:
			res := provider.ServerResponse{HTTPStatusCode: e.Code, Header: e.Header}

			if e.Code == http.StatusConflict || e.Code == http.StatusNotModified {
				return res, &provider.DNSRecordExists{Record: srvRecord.Name}
			}

			return res, e
		default:
			return provider.ServerResponse{}, e
		}

	}

	return provider.ServerResponse{HTTPStatusCode: res.HTTPStatusCode, Header: res.Header}, nil
}

func (c *GoogleDnsClient) RemoveExternalDns(hostname string, gs *agonesv1.GameServer) (provider.ServerResponse, error) {
	change := dns.Change{}

	aRecord := NewARecordSet(hostname, gs.Status.Address, gs.Name, DefaultTtl)
	srvRecord := NewSrvRecordSet(hostname, gs, DefaultTtl)

	change.Deletions = []*dns.ResourceRecordSet{srvRecord, aRecord}
	res, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()
	if err != nil {

		switch e := err.(type) {
		case *googleapi.Error:
			return provider.ServerResponse{HTTPStatusCode: e.Code, Header: e.Header},
				&provider.DNSRecordNonExistent{Records: []string{aRecord.Name, srvRecord.Name}, ServerError: e}
		default:
			return provider.ServerResponse{}, e
		}
	}

	return provider.ServerResponse{HTTPStatusCode: res.HTTPStatusCode, Header: res.Header}, nil
}

func (c *GoogleDnsClient) SetNodeExternalDns(hostname string, node *corev1.Node) error {
	change := dns.Change{}

	externalIp, err := scheme.GetNodeExternalAddress(node)
	if err != nil {
		return err
	}

	aRecord := NewARecordSet(hostname, externalIp, node.Name, DefaultTtl)

	change.Additions = []*dns.ResourceRecordSet{aRecord}

	if _, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do(); err != nil {
		return err
	}

	return nil
}

func (c *GoogleDnsClient) RemoveNodeExternalDns(hostname string, node *corev1.Node) error {
	change := dns.Change{}

	externalIp, err := scheme.GetNodeExternalAddress(node)
	if err != nil {
		return err
	}

	aRecord := NewARecordSet(hostname, externalIp, node.Name, DefaultTtl)

	change.Deletions = []*dns.ResourceRecordSet{aRecord}

	if _, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do(); err != nil {
		return err
	}

	return nil
}

func NewSrvRecordSet(hostname string, gs *agonesv1.GameServer, ttl int64) *dns.ResourceRecordSet {

	port := gs.Status.Ports[0].Port

	srvRecordName := mcDns.JoinSrvRecordName(hostname, gs.Name)
	aRecordName := mcDns.JoinARecordName(hostname, gs.Name)

	resourceRecord := mcDns.JoinSrvRR(aRecordName, uint16(port), DefaultPriority, DefaultWeight)

	return &dns.ResourceRecordSet{Type: SRV, Name: srvRecordName, Rrdatas: []string{resourceRecord}, Ttl: ttl}
}

func NewARecordSet(hostname string, hostExternalIp string, resourceName string, ttl int64) *dns.ResourceRecordSet {
	recordName := mcDns.JoinARecordName(hostname, resourceName)
	return &dns.ResourceRecordSet{Type: A, Name: recordName, Rrdatas: []string{hostExternalIp}, Ttl: ttl}
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
