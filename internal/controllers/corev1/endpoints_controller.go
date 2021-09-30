// Copyright (c) 2021 The BFE Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package corev1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/filter"
)

// EndpointsReconciler reconciles a Endpoints object
type EndpointsReconciler struct {
	BfeConfigBuilder *bfeConfig.ConfigBuilder

	client.Client
	Scheme *runtime.Scheme
}

func (r *EndpointsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).Info("reconciling endpoints", "api version", "corev1")

	ep := &corev1.Endpoints{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, ep)
	if err != nil {
		return ctrl.Result{}, nil
	}

	svc := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, svc)
	if err != nil {
		return ctrl.Result{}, nil
	}

	r.BfeConfigBuilder.UpdateService(svc, ep)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EndpointsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Endpoints{}, builder.WithPredicates(filter.NamespaceFilter())).
		Complete(r)
}
