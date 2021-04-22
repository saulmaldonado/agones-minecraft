package main

import (
	"context"
	"log"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func main() {

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up controller")

	ctrl, err := controller.New("mc-controller", mgr, controller.Options{
		Reconciler: reconcile.Func(func(c context.Context, r reconcile.Request) (reconcile.Result, error) {
			log.Println(r.NamespacedName)

			return reconcile.Result{}, nil
		}),
	})
	if err != nil {
		log.Fatal(err)
	}
	err = ctrl.Watch(&source.Kind{Type: &agonesv1.GameServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting manager")

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatal(err)
	}
}
