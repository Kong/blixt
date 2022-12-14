package controllers

import (
	"context"

	"github.com/kong/blixt/pkg/vars"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// mapDataPlaneDaemonsetToUDPRoutes is a mapping function to map dataplane
// DaemonSet updates to UDPRoute reconcilations. This enables changes to the
// DaemonSet such as adding new Pods for a new Node to result in new dataplane
// instances getting fully configured.
func (r *UDPRouteReconciler) mapDataPlaneDaemonsetToUDPRoutes(obj client.Object) (reqs []reconcile.Request) {
	daemonset, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return
	}

	// determine if this is a blixt daemonset
	matchLabels := daemonset.Spec.Selector.MatchLabels
	app, ok := matchLabels["app"]
	if !ok || app != vars.DefaultDataPlaneAppLabel {
		return
	}

	// verify that it's the dataplane daemonset
	component, ok := matchLabels["component"]
	if !ok || component != vars.DefaultDataPlaneComponentLabel {
		return
	}

	udproutes := &gatewayv1alpha2.UDPRouteList{}
	ctx := context.Background()
	if err := r.Client.List(ctx, udproutes); err != nil {
		// TODO: https://github.com/kubernetes-sigs/controller-runtime/issues/1996
		log := log.FromContext(ctx)
		log.Error(err, "could not enqueue UDPRoutes for DaemonSet update")
		return
	}

	for _, udproute := range udproutes.Items {
		reqs = append(reqs, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: udproute.Namespace,
				Name:      udproute.Name,
			},
		})
	}

	return
}
