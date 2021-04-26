package main

import (
	"flag"
	"os"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	ctrl "github.com/saulmaldonado/agones-minecraft/controller/internal/controller"
	schm "github.com/saulmaldonado/agones-minecraft/controller/internal/controller/scheme"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider/google"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	ManagedZone  string
	ProjectID    string
	NodeHostname string
)

func init() {
	flag.StringVar(&ManagedZone, "zone", "", "DNS zone that the controller will manage")
	flag.StringVar(&ProjectID, "project", "", "Google project ID")

	flag.Parse()
}

func main() {
	scheme := runtime.NewScheme()
	log := zap.New()

	log.Info("Adding scheme")

	if err := schm.AddToScheme(scheme); err != nil {
		log.Error(err, "Error adding scheme")
		os.Exit(1)
	}

	log.Info("Initializing DNS client")

	dns, err := google.NewDnsClient(ProjectID, ManagedZone)

	if err != nil {
		log.Error(err, "Error Initializing DNS client")
		os.Exit(1)
	}

	log.Info("Setting up manager")

	manager, err := controller.NewManager(config.GetConfigOrDie(), controller.Options{
		Scheme: scheme,
		Logger: log,
	})
	if err != nil {
		log.Error(err, "Error setting up manager")
		os.Exit(1)
	}

	log.Info("Setting up GameServer controller")

	if err = controller.NewControllerManagedBy(manager).
		For(&agonesv1.GameServer{}).WithEventFilter(predicate.NewPredicateFuncs(func(object client.Object) bool {
		gs := object.(*agonesv1.GameServer)
		return !schm.IsBeforePodCreated(gs)
	})).
		Complete(ctrl.NewGameServerReconciler(manager.GetClient(), manager.GetScheme(), log, dns)); err != nil {

		log.Error(err, "Error setting up GameServer controller")
		os.Exit(1)
	}

	log.Info("Setting up Node controller")

	if err := controller.NewControllerManagedBy(manager).
		For(&corev1.Node{}).
		Complete(ctrl.NewNodeReconciler(manager.GetClient(), manager.GetScheme(), log, dns)); err != nil {

		log.Error(err, "Error setting up Node controller")
		os.Exit(1)
	}

	log.Info("Starting manager")
	if err := manager.Start(controller.SetupSignalHandler()); err != nil {
		log.Error(err, "Error starting manager")
		os.Exit(1)
	}
}
