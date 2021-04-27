package google

import (
	"context"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"cloud.google.com/go/compute/metadata"
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
	AlreadyExists   string = "alreadyExists"
)

func (c *GoogleDnsClient) SetGameServerExternalDns(hostname string, gs *agonesv1.GameServer) error {
	change := dns.Change{}

	nodeARecord := mcDns.JoinARecordName(hostname, gs.Status.NodeName)
	srvRecord := NewSrvRecordSet(hostname, gs, DefaultTtl, nodeARecord)

	change.Additions = []*dns.ResourceRecordSet{srvRecord}
	_, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	return err
}

func (c *GoogleDnsClient) RemoveGameServerExternalDns(hostname string, gs *agonesv1.GameServer) error {
	change := dns.Change{}

	nodeARecord := mcDns.JoinARecordName(hostname, gs.Status.NodeName)
	srvRecord := NewSrvRecordSet(hostname, gs, DefaultTtl, nodeARecord)

	change.Deletions = []*dns.ResourceRecordSet{srvRecord}
	_, err := c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	return err
}

func (c *GoogleDnsClient) SetNodeExternalDns(hostname string, node *corev1.Node) error {
	change := dns.Change{}

	externalIp, err := scheme.GetNodeExternalAddress(node)
	if err != nil {
		return err
	}

	aRecord := NewARecordSet(hostname, externalIp, node.Name, DefaultTtl)

	change.Additions = []*dns.ResourceRecordSet{aRecord}

	_, err = c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	return err
}

func (c *GoogleDnsClient) RemoveNodeExternalDns(hostname string, node *corev1.Node) error {
	change := dns.Change{}

	externalIp, err := scheme.GetNodeExternalAddress(node)
	if err != nil {
		return err
	}

	aRecord := NewARecordSet(hostname, externalIp, node.Name, DefaultTtl)

	change.Deletions = []*dns.ResourceRecordSet{aRecord}

	_, err = c.Changes.Create(c.config.GoogleProjectId, c.config.GoogleManagedZone, &change).Do()

	return err
}

func NewSrvRecordSet(hostname string, gs *agonesv1.GameServer, ttl int64, aRecordName string) *dns.ResourceRecordSet {
	port := gs.Status.Ports[0].Port

	srvRecordName := mcDns.JoinSrvRecordName(hostname, gs.Name)

	resourceRecord := mcDns.JoinSrvRR(aRecordName, uint16(port), DefaultPriority, DefaultWeight, aRecordName)

	return &dns.ResourceRecordSet{Type: SRV, Name: srvRecordName, Rrdatas: []string{resourceRecord}, Ttl: ttl}
}

func NewARecordSet(hostname string, hostExternalIp string, resourceName string, ttl int64) *dns.ResourceRecordSet {
	recordName := mcDns.JoinARecordName(hostname, resourceName)
	return &dns.ResourceRecordSet{Type: A, Name: recordName, Rrdatas: []string{hostExternalIp}, Ttl: ttl}
}

func (c *GoogleDnsClient) IgnoreClientError(err error) error {
	if _, ok := err.(*googleapi.Error); ok {
		return nil
	}
	return err
}

func (c *GoogleDnsClient) IgnoreAlreadyExists(err error) error {
	if apiErr, ok := err.(*googleapi.Error); ok {
		for _, e := range apiErr.Errors {
			if e.Reason != AlreadyExists {
				return err
			}
		}
	}

	return nil
}

func NewDnsClient(managedZone, projectId string) (*GoogleDnsClient, error) {
	gcloud, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, err
	}

	dns, err := dns.NewService(context.Background(), option.WithHTTPClient(gcloud))

	if err != nil {
		return nil, err
	}

	if projectId == "" {
		GCEProjectId, err := metadata.ProjectID()
		if err != nil {
			return nil, err
		}
		projectId = GCEProjectId
	}

	config := provider.Config{GoogleProjectId: projectId, GoogleManagedZone: managedZone}

	return &GoogleDnsClient{config: config, Service: dns}, nil
}
