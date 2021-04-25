package controller

import (
	"context"
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-logr/logr"
	mcDns "github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	AnnotationPrefix      string = "agones-mc"
	HostnameAnnotation    string = "hostname"
	ExternalDnsAnnotation string = "externalDNS"
)

type GameServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	log    logr.Logger
	dns    provider.DnsClient
}

func (r *GameServerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	gs := agonesv1.GameServer{}

	if err := r.getGameServer(ctx, req.NamespacedName, &gs); err != nil {
		return reconcile.Result{}, err
	}

	exists := findeExternalDnsAnnotation(&gs)
	hostname, hostnameFound := getHostnameAnnotation(&gs)

	if exists {
		if isGameServerDeleted(&gs) {
			if err := r.cleanUpGameServer(hostname, &gs); err != nil {
				return reconcile.Result{}, err
			}
		}

		r.log.Info(fmt.Sprintf("External DNS set for %s", gs.Name))
		return reconcile.Result{}, nil
	}

	if hostnameFound {
		if err := r.getGameServer(ctx, req.NamespacedName, &gs); err != nil {
			return reconcile.Result{}, err
		}

		if err := r.setGameServerDNS(ctx, hostname, &gs); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	r.log.Info(fmt.Sprintf("No hostname annotation for %s", gs.Name))
	return reconcile.Result{}, nil
}

func (r *GameServerReconciler) setGameServerDNS(ctx context.Context, hostname string, gs *agonesv1.GameServer) error {
	if res, err := r.dns.SetExternalDns(hostname, gs); err != nil {

		switch err.(type) {
		case *provider.DNSRecordExists:
			r.log.Info(err.Error(), "HTTPStatusCode", res.HTTPStatusCode)
		default:
			r.log.Error(err, "Error creating DNS records")
			return err
		}

	}

	externalDns := setExternalDnsAnnotation(mcDns.JoinARecordName(hostname, gs.Name), gs)

	if err := r.Update(ctx, gs); err != nil {
		return err
	}

	r.log.Info(fmt.Sprintf("GameServer %s externalDNS set to %s", gs.Name, externalDns))
	return nil
}

func (r *GameServerReconciler) getGameServer(ctx context.Context, namespacedName types.NamespacedName, gs *agonesv1.GameServer) error {
	if err := r.Get(ctx, namespacedName, gs); err != nil {

		if errors.IsNotFound(err) {
			r.log.Info(fmt.Sprintf("Could not find GameServer %s", namespacedName))
			return err
		}

		r.log.Error(err, "Error getting GameServer")
		return err
	}

	return nil
}

func isGameServerDeleted(gs *agonesv1.GameServer) bool {
	return !gs.ObjectMeta.DeletionTimestamp.IsZero()
}

func (r *GameServerReconciler) cleanUpGameServer(hostname string, gs *agonesv1.GameServer) error {
	if res, err := r.dns.RemoveExternalDns(hostname, gs); err != nil {
		switch e := err.(type) {
		case *provider.DNSRecordNonExistent:
			r.log.Info(err.Error(), "HTTPStatusCode", res.HTTPStatusCode, "ServerError", e.ServerError)
		default:
			return err
		}
	}

	r.log.Info(fmt.Sprintf("GameServer %s externalDNS records removed", gs.Name))
	return nil
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) reconcile.Reconciler {
	return &GameServerReconciler{client, scheme, log, dns}
}
