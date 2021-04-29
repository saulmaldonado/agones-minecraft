package controller_test

import (
	"path/filepath"
	"testing"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/controller"
	schm "github.com/saulmaldonado/agones-minecraft/controller/internal/controller/scheme"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	testClient      client.Client
	testEnv         *envtest.Environment
	FakeDns         *TestDnsClient
	DefaultPriority = 0
	DefaultWeight   = 0
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller suite")
}

var _ = BeforeSuite(func() {

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstraping test env")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "crd")},
	}

	config, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(config).NotTo(BeNil())

	agonesv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	testClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(testClient).NotTo(BeNil())

	manager, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme.Scheme,
	})

	Expect(err).NotTo(HaveOccurred())

	FakeDns = &TestDnsClient{DnsRecords: []string{}}

	err = ctrl.NewControllerManagedBy(manager).For(&agonesv1.GameServer{}).WithEventFilter(
		predicate.NewPredicateFuncs(func(object client.Object) bool {
			gs := object.(*agonesv1.GameServer)
			return !schm.IsBeforePodCreated(gs)
		})).Complete(
		&controller.GameServerReconciler{
			DnsReconciler: controller.DnsReconciler{
				Client: manager.GetClient(),
				Scheme: manager.GetScheme(),
				Log:    ctrl.Log.WithName("controllers").WithName("GameServer"),
				Dns:    FakeDns,
			},
		},
	)

	Expect(err).NotTo(HaveOccurred())

	go func() {
		manager.Start(ctrl.SetupSignalHandler())
		Expect(err).NotTo(HaveOccurred())
	}()
}, 60)

type TestDnsClient struct {
	DnsRecords []string
}

func (d *TestDnsClient) SetGameServerExternalDns(hostname string, gs *agonesv1.GameServer) error {
	nodeARecord := dns.JoinARecordName(hostname, gs.Status.NodeName)
	srvRecord := dns.JoinSrvRecordName(hostname, gs.Name)

	port := gs.Status.Ports[0].Port

	resourceRecord := dns.JoinSrvRR(srvRecord, uint16(port), DefaultPriority, DefaultWeight, nodeARecord)

	d.DnsRecords = append(d.DnsRecords, srvRecord+" "+resourceRecord)

	return nil
}

func (d *TestDnsClient) RemoveGameServerExternalDns(hostname string, gs *agonesv1.GameServer) error {
	nodeARecord := dns.JoinARecordName(hostname, gs.Status.NodeName)
	srvRecord := dns.JoinSrvRecordName(hostname, gs.Name)

	port := gs.Status.Ports[0].Port

	resourceRecord := dns.JoinSrvRR(srvRecord, uint16(port), DefaultPriority, DefaultWeight, nodeARecord)

	recordToDelete := srvRecord + " " + resourceRecord

	for i, record := range d.DnsRecords {
		if record == recordToDelete {
			d.DnsRecords = append(d.DnsRecords[:i], d.DnsRecords[i+1:]...)
		}
	}

	return nil
}

func (d *TestDnsClient) SetNodeExternalDns(hostname string, node *corev1.Node) error {
	return nil
}

func (*TestDnsClient) RemoveNodeExternalDns(hostname string, node *corev1.Node) error {
	return nil
}

func (*TestDnsClient) IgnoreClientError(err error) error {
	return nil
}

func (*TestDnsClient) IgnoreAlreadyExists(err error) error {
	return nil
}

var _ = AfterSuite(func() {
	By("tearing down the test env")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
