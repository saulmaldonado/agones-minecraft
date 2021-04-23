package main

import (
	"context"
	"fmt"
	"log"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type GameServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func main() {
	scheme := runtime.NewScheme()
	agonesv1.AddToScheme(scheme)
	clientGoScheme.AddToScheme(scheme)

	log.Println("Setting up controller")

	manager, err := controller.NewManager(config.GetConfigOrDie(), controller.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = controller.NewControllerManagedBy(manager).
		For(&agonesv1.GameServer{}).
		Complete(&GameServerReconciler{manager.GetClient(), manager.GetScheme()})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting manager")
	if err := manager.Start(controller.SetupSignalHandler()); err != nil {
		log.Fatal(err)
	}
}

func (r *GameServerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	gs := &agonesv1.GameServer{}

	err := r.Get(ctx, req.NamespacedName, gs)
	if errors.IsNotFound(err) {
		log.Println("could not find GameServer")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not get GameServer")
	}

	if gs.Annotations["agones-mc/hostname"] != "" && gs.Annotations["agones-mc/externalDNS"] == "" {
		log.Println("Adding external DNS")
		hostname := gs.Annotations["agones-mc/hostname"]
		gs.Annotations["agones-mc/externalDNS"] = gs.Name + "." + hostname
		if err := r.Update(ctx, gs); err != nil {
			return reconcile.Result{}, err
		}
		log.Println("GameServer updated")
	}

	return reconcile.Result{}, nil
}
