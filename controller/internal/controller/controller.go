package controller

import (
	"context"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-logr/logr"
	schm "github.com/saulmaldonado/agones-minecraft/controller/internal/controller/scheme"
	mcDns "github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DnsReconciler struct {
	client.Client
	scheme *runtime.Scheme
	log    logr.Logger
	dns    provider.DnsClient
}

func (r *DnsReconciler) ReconcileDns(ctx context.Context, req reconcile.Request, obj client.Object) (reconcile.Result, error) {
	if err := r.getResource(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	dnsExists := findExternalDnsAnnotation(obj)
	domain, domainFound := getDomainAnnotationOrLabel(obj)

	if dnsExists {
		if schm.IsResourceDeleted(obj) {

			if err := r.cleanUpResource(domain, obj); err != nil {
				r.log.Error(err, "Error cleaning up resource DNS", "Resource", schm.GVKString(obj), "Name", obj.GetName())
			} else {
				r.log.Info("DNS record removed", "Resource", schm.GVKString(obj), "Name", obj.GetName())
			}
		}

		return reconcile.Result{}, nil
	}

	if domainFound {
		if err := r.getResource(ctx, req.NamespacedName, obj); err != nil {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}

		if err := r.setupResource(ctx, domain, obj); err != nil {
			r.log.Error(err, "Error setting Resource DNS", "Resource", schm.GVKString(obj), "Name", obj.GetName())
			return reconcile.Result{}, r.dns.IgnoreClientError(err)
		}

		r.log.Info("New DNS record set", "Resource", schm.GVKString(obj), "Name", obj.GetName())
		return reconcile.Result{}, nil
	}

	r.log.Info("No domain annotation or label", "Resource", schm.GVKString(obj), "Name", obj.GetName())
	return reconcile.Result{}, nil
}

func (r *DnsReconciler) getResource(ctx context.Context, namespacedName types.NamespacedName, obj client.Object) error {
	err := r.Get(ctx, namespacedName, obj)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Could not find resource", "Name", namespacedName.String())
	}

	return err
}

func (r *DnsReconciler) cleanUpResource(hostname string, obj client.Object) error {
	switch res := obj.(type) {
	case *agonesv1.GameServer:
		return r.dns.RemoveGameServerExternalDns(hostname, res)
	case *corev1.Node:
		return r.dns.RemoveNodeExternalDns(hostname, res)
	}

	return nil
}

func (r *DnsReconciler) setupResource(ctx context.Context, hostname string, obj client.Object) error {
	var err error

	switch res := obj.(type) {
	case *agonesv1.GameServer:
		err = r.dns.SetGameServerExternalDns(hostname, res)
	case *corev1.Node:
		err = r.dns.SetNodeExternalDns(hostname, res)
	}

	if err != nil {
		return err
	}

	setExternalDnsAnnotation(mcDns.JoinARecordName(hostname, obj.GetName()), obj)

	if err := r.Update(ctx, obj); err != nil {
		return err
	}

	return nil
}
