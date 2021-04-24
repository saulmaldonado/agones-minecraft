package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type GameServerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	log       logr.Logger
	dns       *dns.Service
	cfg       *GoogleConfig
	endpoints map[string]bool
	lock      sync.Mutex
}

type GoogleConfig struct {
	ProjectId   string
	ManagedZone string
}

var (
	ManagedZone string
	ProjectID   string
)

func init() {
	flag.StringVar(&ManagedZone, "zone", "", "DNS zone that the controller will manage")
	flag.StringVar(&ProjectID, "project", "", "Google project ID")

	flag.Parse()
}

func main() {
	scheme := runtime.NewScheme()
	log := zap.New()

	if err := agonesv1.AddToScheme(scheme); err != nil {
		log.Error(err, "Error Adding agones.dev/v1 to scheme")
	}

	if err := clientGoScheme.AddToScheme(scheme); err != nil {
		log.Error(err, "Error adding client-go clientset to scheme")
	}

	log.Info("Initializing DNS client")

	gcloud, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
	if err != nil {
		log.Error(err, "Error getting credentials")
	}

	dns, err := dns.NewService(context.Background(), option.WithHTTPClient(gcloud))
	if err != nil {
		log.Error(err, "Error getting DNS client")
	}

	log.Info("Setting up manager")

	manager, err := controller.NewManager(config.GetConfigOrDie(), controller.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Error(err, "Error setting up manager")
		os.Exit(1)
	}

	log.Info("Setting up controller")

	cfg := GoogleConfig{ProjectID, ManagedZone}

	fmt.Println(cfg)

	err = controller.NewControllerManagedBy(manager).
		For(&agonesv1.GameServer{}).
		Complete(&GameServerReconciler{manager.GetClient(), manager.GetScheme(), log, dns, &cfg, map[string]bool{}, sync.Mutex{}})
	if err != nil {
		log.Error(err, "Error setting up controller")
		os.Exit(1)
	}

	log.Info("Starting manager")
	if err := manager.Start(controller.SetupSignalHandler()); err != nil {
		log.Error(err, "Error starting manager")
		os.Exit(1)
	}
}

func (r *GameServerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	gs := &agonesv1.GameServer{}

	err := r.Get(ctx, req.NamespacedName, gs)
	if errors.IsNotFound(err) {
		r.log.Error(err, "could not find GameServer")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not get GameServer")
	}

	if gs.Annotations["agones-mc/hostname"] != "" && gs.Annotations["agones-mc/externalDNS"] == "" {
		hostname := gs.Annotations["agones-mc/hostname"]

		r.Get(ctx, req.NamespacedName, gs)

		if err := r.setExternalDNS(hostname, gs); err != nil {

			switch err := err.(type) {
			case *RecordExistsErr:
				r.log.Info("Record already exists")
			case *ServerNotReady:
				r.log.Info("Server not ready")
			default:
				return reconcile.Result{}, err
			}

		}

		gs.Annotations["agones-mc/externalDNS"] = gs.Name + "." + hostname
		if err := r.Update(ctx, gs); err != nil {
			return reconcile.Result{}, err
		}
		r.log.Info("GameServer updated")
	}

	return reconcile.Result{}, nil
}

func (r *GameServerReconciler) setExternalDNS(hostname string, gs *agonesv1.GameServer) error {
	if gs.Status.Address == "" || len(gs.Status.Ports) == 0 {
		return &ServerNotReady{}
	}

	hostExternalIp := gs.Status.Address
	port := gs.Status.Ports[0].Port

	r.lock.Lock()
	defer r.lock.Unlock()

	recordName := fmt.Sprintf("%s.%s.", gs.Name, hostname)
	srvRecordName := fmt.Sprintf("_minecraft._tcp.%s", recordName)

	if _, ok := r.endpoints[recordName]; ok {
		return &RecordExistsErr{recordName}
	}

	change := dns.Change{}

	srvResourceRecord := fmt.Sprintf("0 1 %d %s", port, recordName)

	srvRecord := dns.ResourceRecordSet{Type: "SRV", Name: srvRecordName, Rrdatas: []string{srvResourceRecord}, Ttl: 60 * 30}
	aRecord := dns.ResourceRecordSet{Type: "A", Name: recordName, Rrdatas: []string{hostExternalIp}, Ttl: 60 * 30}

	change.Additions = []*dns.ResourceRecordSet{&aRecord, &srvRecord}

	res, err := r.dns.Changes.Create(r.cfg.ProjectId, r.cfg.ManagedZone, &change).Do()

	if err != nil {
		return err
	}

	r.endpoints[recordName] = true

	json, _ := res.MarshalJSON()

	log.Println(string(json))

	return nil
}

type RecordExistsErr struct {
	record string
}

func (e *RecordExistsErr) Error() string {
	return fmt.Sprintf("Record %s already exists", e.record)
}

type ServerNotReady struct{}

func (e *ServerNotReady) Error() string {
	return "Server IP and port not allocated"
}
