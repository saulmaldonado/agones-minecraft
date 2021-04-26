package controller

import (
	"context"

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

// type GameServerReconciler struct {
// 	client.Client
// 	Scheme *runtime.Scheme
// 	log    logr.Logger
// 	dns    provider.DnsClient
// }

// func (r *GameServerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
// 	gs := agonesv1.GameServer{}

// 	if err := r.getGameServer(ctx, req.NamespacedName, &gs); err != nil {
// 		return reconcile.Result{}, nil
// 	}

// 	if !r.gameServerIsReady(&gs) {
// 		return reconcile.Result{}, nil
// 	}

// 	exists := findExternalDnsAnnotation(&gs)
// 	hostname, hostnameFound := getHostnameAnnotation(&gs)

// 	if exists {
// 		if isResourceDeleted(&gs) {
// 			if err := r.cleanUpGameServer(hostname, &gs); err != nil {
// 				return reconcile.Result{}, err
// 			}

// 			return reconcile.Result{}, nil
// 		}

// 		r.log.Info(fmt.Sprintf("ExternalDNS set for %s", gs.Name))
// 		return reconcile.Result{}, nil
// 	}

// 	if hostnameFound {
// 		if err := r.getGameServer(ctx, req.NamespacedName, &gs); err != nil {
// 			return reconcile.Result{}, err
// 		}

// 		if err := r.setGameServerDNS(ctx, hostname, &gs); err != nil {
// 			return reconcile.Result{}, err
// 		}

// 		return reconcile.Result{}, nil
// 	}

// 	r.log.Info(fmt.Sprintf("No hostname annotation for %s", gs.Name))
// 	return reconcile.Result{}, nil
// }

// func (r *GameServerReconciler) setGameServerDNS(ctx context.Context, hostname string, gs *agonesv1.GameServer) error {
// 	if res, err := r.dns.SetExternalDns(hostname, gs); err != nil {

// 		switch err.(type) {
// 		case *provider.DNSRecordExists:
// 			r.log.Info(err.Error(), "HTTPStatusCode", res.HTTPStatusCode)
// 		default:
// 			r.log.Error(err, "Error creating DNS records")
// 			return err
// 		}

// 	}

// 	externalDns := setExternalDnsAnnotation(mcDns.JoinARecordName(hostname, gs.Name), gs)

// 	if err := r.Update(ctx, gs); err != nil {
// 		return err
// 	}

// 	r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS set to %s", gs.Name, externalDns))
// 	return nil
// }

// func (r *GameServerReconciler) getGameServer(ctx context.Context, namespacedName types.NamespacedName, gs *agonesv1.GameServer) error {
// 	if err := r.Get(ctx, namespacedName, gs); err != nil {

// 		if errors.IsNotFound(err) {
// 			r.log.Info(fmt.Sprintf("Could not find GameServer %s", namespacedName))
// 			return err
// 		}

// 		r.log.Error(err, "Error getting GameServer")
// 		return err
// 	}

// 	return nil
// }

// func (r *GameServerReconciler) gameServerIsReady(gs *agonesv1.GameServer) bool {
// 	if gs.Status.Address == "" && len(gs.Status.Ports) == 0 {
// 		r.log.Info("GameServer not ready to set ExternalDNS", "GameServer name", gs.Name)
// 		return false
// 	}

// 	return true
// }

// func (r *GameServerReconciler) cleanUpGameServer(hostname string, gs *agonesv1.GameServer) error {
// 	if res, err := r.dns.RemoveExternalDns(hostname, gs); err != nil {
// 		switch e := err.(type) {
// 		case *provider.DNSRecordNonExistent:
// 			r.log.Info(err.Error(), "HTTPStatusCode", res.HTTPStatusCode, "ServerError", e.ServerError)
// 		default:
// 			return err
// 		}
// 	}

// 	r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS records removed", gs.Name))
// 	return nil
// }

// func NewReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) reconcile.Reconciler {
// 	return &GameServerReconciler{client, scheme, log, dns}
// }

// type NodeReconciler struct {
// 	client.Client
// 	Scheme *runtime.Scheme
// 	log    logr.Logger
// 	dns    provider.DnsClient
// }

// func NewNodeReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) reconcile.Reconciler {
// 	return &NodeReconciler{client, scheme, log, dns}
// }

// func (r *NodeReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
// 	node := corev1.Node{}

// 	if err := r.getNode(ctx, req.NamespacedName, &node); err != nil {
// 		return reconcile.Result{}, client.IgnoreNotFound(err)
// 	}

// 	exists := findExternalDnsAnnotation(&node)
// 	hostname, hostnameFound := getHostnameAnnotation(&node)

// 	if exists {
// 		if isResourceDeleted(&node) {
// 			if err := r.cleanUpNode(hostname, &node); err != nil {
// 				r.log.Error(err, "Error removing externalDNS")
// 			}

// 			return reconcile.Result{}, nil
// 		}

// 		r.log.Info("ExternalDNS already set", "Node", node.Name)
// 		return reconcile.Result{}, nil
// 	}

// 	if hostnameFound {
// 		if err := r.getNode(ctx, req.NamespacedName, &node); err != nil {
// 			return reconcile.Result{}, client.IgnoreNotFound(err)
// 		}

// 		if err := r.dns.SetNodeExternalDns(hostname, &node); err != nil {
// 			r.log.Error(err, "Error setting Node ExternalDNS")
// 			return reconcile.Result{}, nil
// 		}

// 		externalDns := mcDns.JoinARecordName(hostname, node.GetName())

// 		setExternalDnsAnnotation(externalDns, &node)

// 		if err := r.updateNode(ctx, &node); err != nil {
// 			return reconcile.Result{}, err
// 		}

// 		r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS set to %s", node.Name, externalDns))
// 		return reconcile.Result{}, nil
// 	}

// 	r.log.Info("No hostname annotation", "Node", node.Name)
// 	return reconcile.Result{}, nil
// }

// func (r *NodeReconciler) getNode(ctx context.Context, namespacedName types.NamespacedName, node *corev1.Node) error {
// 	if err := r.Get(ctx, namespacedName, node); err != nil {

// 		if errors.IsNotFound(err) {
// 			r.log.Info(fmt.Sprintf("Could not find Node %s", namespacedName))
// 		}

// 		return err
// 	}

// 	return nil
// }

// func (r *NodeReconciler) updateNode(ctx context.Context, node *corev1.Node) error {
// 	return r.Update(ctx, node)
// }

// func (r *NodeReconciler) cleanUpNode(hostname string, node *corev1.Node) error {
// 	if err := r.dns.RemoveNodeExternalDns(hostname, node); err != nil {
// 		return err
// 	}

// 	r.log.Info(fmt.Sprintf("GameServer %s ExternalDNS records removed", node.Name))
// 	return nil
// }

// type ResourceDnsReconciler interface {
// 	cleanUpResource(hostname string, obj client.Object) error
// 	setUpResource(ctx context.Context, hostname string, obj client.Object) error
// }

type DnsReconciler struct {
	client.Client
	scheme *runtime.Scheme
	log    logr.Logger
	dns    provider.DnsClient
}

func (r *DnsReconciler) getResource(ctx context.Context, namespacedName types.NamespacedName, obj client.Object) error {
	err := r.Get(ctx, namespacedName, obj)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Could not find resource", "Recource", obj.GetObjectKind().GroupVersionKind().String(), "Name", obj.GetName())
	}

	return err
}

func (r *DnsReconciler) ReconcileDns(ctx context.Context, req reconcile.Request, obj client.Object) (reconcile.Result, error) {
	if err := r.getResource(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	exists := findExternalDnsAnnotation(obj)
	hostname, hostnameFound := getHostnameAnnotation(obj)

	if exists {
		if isResourceDeleted(obj) {

			if err := r.cleanUpResource(hostname, obj); err != nil {
				r.log.Error(err, "Error cleaning up resource DNS", "Resource", obj.GetObjectKind().GroupVersionKind().String(), "Name", obj.GetName())
			}

			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, nil
	}

	if hostnameFound {
		if err := r.getResource(ctx, req.NamespacedName, obj); err != nil {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}

		if err := r.setupResource(ctx, hostname, obj); err != nil {
			return reconcile.Result{}, r.dns.IgnoreClientError(err)
		}

		r.log.Info("New DNS record set", "Resource", obj.GetObjectKind().GroupVersionKind().String(), "Name", obj.GetName())
		return reconcile.Result{}, nil
	}

	r.log.Info("No hostname annotation", "Resource", obj.GetObjectKind().GroupVersionKind().String(), "Name", obj.GetName())
	return reconcile.Result{}, nil
}

func isResourceDeleted(obj client.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

// type GameServerReconciler struct {
// 	client.Client
// 	log logr.Logger
// 	dns provider.DnsClient
// }

// func (r *GameServerReconciler) cleanUpResource(hostname string, obj client.Object) error {
// 	gs := obj.(*agonesv1.GameServer)
// }

// func (r *GameServerReconciler) setUpResource(ctx context.Context, hostname string, obj client.Object) error {
// 	gs := obj.(*agonesv1.GameServer)
// 	if err := r.dns.SetGameServerExternalDns(hostname, gs); err != nil {
// 		r.log.Error(err, "Error setting GameServer DNS", "Name", gs.Name)
// 		return err
// 	}

// 	setExternalDnsAnnotation(mcDns.JoinARecordName(hostname, gs.Name), obj)

// 	if err := r.Update(ctx, gs); err != nil {
// 		return err
// 	}

// 	return nil
// }

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

	if r.dns.IgnoreAlreadyExists(err) != nil {
		r.log.Error(err, "Error setting Resource DNS", "Resource", obj.GetObjectKind().GroupVersionKind().String(), "Name", obj.GetName())
		return err
	}

	setExternalDnsAnnotation(mcDns.JoinARecordName(hostname, obj.GetName()), obj)
	if err := r.Update(ctx, obj); err != nil {
		return err
	}

	return nil
}

type GameServerReconciler struct {
	DnsReconciler
}

func (r *GameServerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	gs := agonesv1.GameServer{}
	return r.ReconcileDns(ctx, req, &gs)
}

func NewGameServerReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) *GameServerReconciler {
	return &GameServerReconciler{DnsReconciler: DnsReconciler{client, scheme, log, dns}}
}

type NodeReconciler struct {
	DnsReconciler
}

func NewNodeReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) *NodeReconciler {
	return &NodeReconciler{DnsReconciler: DnsReconciler{client, scheme, log, dns}}
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	node := corev1.Node{}
	return r.ReconcileDns(ctx, req, &node)
}

// type NodeDNSReconciler struct {
// 	client.Client
// 	log logr.Logger
// 	dns provider.DnsClient
// }

// func (r *NodeDNSReconciler) cleanUpResource(hostname string, obj client.Object) error {
// 	node := obj.(*corev1.Node)
// 	return r.dns.RemoveNodeExternalDns(hostname, node)
// }

// func (r *NodeDNSReconciler) setUpResource(ctx context.Context, hostname string, obj client.Object) error {
// 	node := obj.(*corev1.Node)
// 	if err := r.dns.SetNodeExternalDns(hostname, node); err != nil {
// 		r.log.Error(err, "Error setting Node DNS", "Name", node.Name)
// 		return err
// 	}

// 	setExternalDnsAnnotation(mcDns.JoinARecordName(hostname, node.Name), obj)

// 	if err := r.Update(ctx, node); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func NewDNSReconciler(client client.Client, scheme *runtime.Scheme, log logr.Logger, dns provider.DnsClient) *DnsReconciler {
// 	return &DnsReconciler{client, scheme, log, dns, &GameServerReconciler{client, log, dns}}
// }
