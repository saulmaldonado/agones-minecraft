package controller

import (
	"context"
	"fmt"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-logr/logr"
	mcDns "github.com/saulmaldonado/agones-minecraft/controller/internal/dns"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	corev1 "k8s.io/api/core/v1"
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
		return reconcile.Result{}, nil
	}

	if !r.gameServerIsReady(&gs) {
		return reconcile.Result{}, nil
	}

	exists := findExternalDnsAnnotation(&gs)
	hostname, hostnameFound := getHostnameAnnotation(&gs)

	if exists {
		if isResourceDeleted(&gs) {
			if err := r.cleanUpGameServer(hostname, &gs); err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}

		r.log.Info(fmt.Sprintf("ExternalDNS set for %s", gs.Name))
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

	r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS set to %s", gs.Name, externalDns))
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

func (r *GameServerReconciler) gameServerIsReady(gs *agonesv1.GameServer) bool {
	if gs.Status.Address == "" && len(gs.Status.Ports) == 0 {
		r.log.Info("GameServer not ready to set ExternalDNS", "GameServer name", gs.Name)
		return false
	}

	return true
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

	r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS records removed", gs.Name))
	return nil
}

func NewReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) reconcile.Reconciler {
	return &GameServerReconciler{client, scheme, log, dns}
}

func isResourceDeleted(obj client.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	log    logr.Logger
	dns    provider.DnsClient
}

func NewNodeReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) reconcile.Reconciler {
	return &NodeReconciler{client, scheme, log, dns}
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	node := corev1.Node{}

	if err := r.getNode(ctx, req.NamespacedName, &node); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	exists := findExternalDnsAnnotation(&node)
	hostname, hostnameFound := getHostnameAnnotation(&node)

	if exists {
		if isResourceDeleted(&node) {
			if err := r.cleanUpNode(hostname, &node); err != nil {
				r.log.Error(err, "Error removing externalDNS")
			}

			return reconcile.Result{}, nil
		}

		r.log.Info("ExternalDNS already set", "Node", node.Name)
		return reconcile.Result{}, nil
	}

	if hostnameFound {
		if err := r.getNode(ctx, req.NamespacedName, &node); err != nil {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}

		if err := r.dns.SetNodeExternalDns(hostname, &node); err != nil {
			r.log.Error(err, "Error setting Node ExternalDNS")
			return reconcile.Result{}, nil
		}

		externalDns := mcDns.JoinARecordName(hostname, node.GetName())

		setExternalDnsAnnotation(externalDns, &node)

		if err := r.updateNode(ctx, &node); err != nil {
			return reconcile.Result{}, err
		}

		r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS set to %s", node.Name, externalDns))
		return reconcile.Result{}, nil
	}

	r.log.Info("No hostname annotation", "Node", node.Name)
	return reconcile.Result{}, nil
}

func (r *NodeReconciler) getNode(ctx context.Context, namespacedName types.NamespacedName, node *corev1.Node) error {
	if err := r.Get(ctx, namespacedName, node); err != nil {

		if errors.IsNotFound(err) {
			r.log.Info(fmt.Sprintf("Could not find Node %s", namespacedName))
		}

		return err
	}

	return nil
}

func (r *NodeReconciler) updateNode(ctx context.Context, node *corev1.Node) error {
	return r.Update(ctx, node)
}

func (r *NodeReconciler) cleanUpNode(hostname string, node *corev1.Node) error {
	if err := r.dns.RemoveNodeExternalDns(hostname, node); err != nil {
		return err
	}

	r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS records removed", node.Name))
	return nil
}
