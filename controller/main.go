package main

import (
	"flag"
	"os"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	ctrl "github.com/saulmaldonado/agones-minecraft/controller/internal/controller"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider/google"
	"k8s.io/apimachinery/pkg/runtime"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

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
		os.Exit(1)
	}

	if err := clientGoScheme.AddToScheme(scheme); err != nil {
		log.Error(err, "Error adding client-go clientset to scheme")
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
	})
	if err != nil {
		log.Error(err, "Error setting up manager")
		os.Exit(1)
	}

	log.Info("Setting up controller")

	err = controller.NewControllerManagedBy(manager).
		For(&agonesv1.GameServer{}).
		Complete(ctrl.NewReconciler(manager.GetClient(), manager.GetScheme(), log, dns))
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
