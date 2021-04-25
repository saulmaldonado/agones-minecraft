package controller

import (
	"context"
	"fmt"
	"strings"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-logr/logr"
	mcDns "github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

	if err := r.Get(ctx, req.NamespacedName, &gs); err != nil {

		if errors.IsNotFound(err) {
			r.log.Info(fmt.Sprintf("Could not find GameServer %s", req.NamespacedName))
			return reconcile.Result{}, nil
		}

		r.log.Error(err, "Error getting GameServer")
		return reconcile.Result{}, err
	}

	if ready := isServerAllocated(&gs); !ready {
		r.log.Info("Waiting for port/address allocation")
		return reconcile.Result{}, nil
	}

	exists := externalDnsExists(&gs)
	hostname, hostnameFound := getHostnameAnnotation(&gs)

	if exists {
		if gs.ObjectMeta.DeletionTimestamp.IsZero() {
			r.log.Info(fmt.Sprintf("External DNS set for %s", gs.Name))
			return reconcile.Result{}, nil
		}

		if res, err := r.dns.RemoveExternalDns(hostname, &gs); err != nil {
			switch e := err.(type) {
			case *provider.DNSRecordNonExistent:
				r.log.Info(err.Error(), "HTTPStatusCode", res.HTTPStatusCode, "ServerError", e.ServerError)
			default:
				r.log.Error(err, fmt.Sprintf("Error deleting DNS record for %s", gs.Name))
				return reconcile.Result{}, err
			}

		}

		r.log.Info(fmt.Sprintf("GameServer %s externalDNS records removed", gs.Name))
		return reconcile.Result{}, nil
	}

	if hostnameFound {
		if res, err := r.dns.SetExternalDns(hostname, &gs); err != nil {

			switch err.(type) {
			case *provider.DNSRecordExists:
				r.log.Info(err.Error(), "HTTPStatusCode", res.HTTPStatusCode)
			default:
				r.log.Error(err, "Error creating DNS records")
				return reconcile.Result{}, err
			}

		}

		r.Get(ctx, req.NamespacedName, &gs)
		externalDns := setExternalDnsAnnotation(mcDns.JoinARecordName(hostname, gs.Name), &gs)

		if err := r.Update(ctx, &gs); err != nil {
			return reconcile.Result{}, err
		}

		r.log.Info(fmt.Sprintf("GameServer %s externalDNS set to %s", gs.Name, externalDns))
		return reconcile.Result{}, nil
	}

	r.log.Info(fmt.Sprintf("No hostname annotation for %s", gs.Name))
	return reconcile.Result{}, nil
}

func externalDnsExists(gs *agonesv1.GameServer) bool {
	_, ok := getAnnotation(ExternalDnsAnnotation, gs)
	return ok
}

func getHostnameAnnotation(gs *agonesv1.GameServer) (string, bool) {
	return getAnnotation(HostnameAnnotation, gs)
}

func setExternalDnsAnnotation(recordName string, gs *agonesv1.GameServer) string {
	setAnnotation(ExternalDnsAnnotation, recordName, gs)
	return recordName
}

func getAnnotation(suffix string, gs *agonesv1.GameServer) (string, bool) {
	annotation := fmt.Sprintf("%s/%s", AnnotationPrefix, suffix)
	hostname, ok := gs.Annotations[annotation]

	if !ok || strings.TrimSpace(hostname) == "" {
		return "", false
	}

	return hostname, true
}

func setAnnotation(suffix string, value string, gs *agonesv1.GameServer) {
	annotation := fmt.Sprintf("%s/%s", AnnotationPrefix, ExternalDnsAnnotation)
	gs.Annotations[annotation] = value
}

func isServerAllocated(gs *agonesv1.GameServer) bool {
	if gs.Status.Address == "" || len(gs.Status.Ports) == 0 {
		return false
	}

	return true
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) reconcile.Reconciler {
	return &GameServerReconciler{client, scheme, log, dns}
}
