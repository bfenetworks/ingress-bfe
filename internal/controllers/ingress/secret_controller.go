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

package ingress

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/filter"
)

func AddSecretController(mgr manager.Manager, cb *bfeConfig.ConfigBuilder) error {
	reconciler := newSecretReconciler(mgr, cb)
	if err := reconciler.setupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create ingress controller")
	}

	return nil
}

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	BfeConfigBuilder *bfeConfig.ConfigBuilder

	client.Client
	Scheme *runtime.Scheme
}

func newSecretReconciler(mgr manager.Manager, cb *bfeConfig.ConfigBuilder) *SecretReconciler {
	return &SecretReconciler{
		BfeConfigBuilder: cb,
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
	}
}

func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).Info("reconciling Secret", "api version", "corev1")

	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, secret)
	if err != nil {
		return ctrl.Result{}, nil
	}

	r.BfeConfigBuilder.UpdateSecret(secret)

	return ctrl.Result{}, nil
}

// setupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) setupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}, builder.WithPredicates(filter.NamespaceFilter())).
		Complete(r)
}
