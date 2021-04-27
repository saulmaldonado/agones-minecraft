package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

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
