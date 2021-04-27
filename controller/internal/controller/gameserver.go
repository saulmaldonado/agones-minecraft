package controller

import (
	"context"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"github.com/go-logr/logr"
	"github.com/saulmaldonado/agones-minecraft/controller/internal/provider"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

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
