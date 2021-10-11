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

package netv1

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/filter"
	"github.com/bfenetworks/ingress-bfe/internal/option"
)

// IngressReconciler reconciles a netv1 Ingress object
type IngressReconciler struct {
	BfeConfigBuilder *bfeConfig.ConfigBuilder

	client.Client
	Scheme *runtime.Scheme
}

func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).Info("reconciling ingress", "api version", "netv1")

	// read ingress
	ingress := &netv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		r.BfeConfigBuilder.DeleteIngress(req.Namespace, req.Name)
		log.V(1).Info("reconcile: ingress delete")
		return reconcile.Result{}, err
	}

	if !filter.MatchIngressClass(ctx, r, ingress.Annotations, ingress.Spec.IngressClassName) {
		return reconcile.Result{}, nil
	}

	log.V(1).Info("reconcile: ingress object", "ingress", ingress)

	err = ReconcileV1Ingress(ctx, r.Client, r.BfeConfigBuilder, ingress)
	setStatus(ctx, r.Client, err, ingress)
	return reconcile.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netv1.Ingress{}, builder.WithPredicates(filter.NamespaceFilter())).
		Complete(r)
}

func setStatus(ctx context.Context, r client.Client, err error, ingress *netv1.Ingress) {
	log := log.FromContext(ctx)

	if annotations.CompareStatus(err, ingress.Annotations[annotations.StatusAnnotationKey]) == 0 {
		// no need to update status if error is not changed
		return
	}

	patch := client.MergeFrom(ingress.DeepCopy())
	ingress.Annotations[annotations.StatusAnnotationKey] = annotations.GenErrorMsg(err)
	if err := r.Patch(ctx, ingress, patch); err != nil {
		log.Error(err, "fail to update annotation")
	}
}

func ReconcileV1Ingress(ctx context.Context, r client.Client, configBuilder *bfeConfig.ConfigBuilder, ingress *netv1.Ingress) error {
	service, endpoints, err := getIngressBackends(ctx, r, ingress)
	if err != nil {
		configBuilder.DeleteIngress(ingress.Namespace, ingress.Name)
		return err
	}

	if len(option.Opts.DefaultBackend) > 0 {
		// use default backend from controller command line argument
		setDefautBackend(ingress, service[option.Opts.DefaultBackend])
	}

	secrets, err := getIngressSecret(ctx, r, ingress)
	if err != nil {
		configBuilder.DeleteIngress(ingress.Namespace, ingress.Name)
		return err
	}

	if err = configBuilder.UpdateIngress(ingress, service, endpoints, secrets); err != nil {
		configBuilder.DeleteIngress(ingress.Namespace, ingress.Name)
		return err
	}

	return nil
}

func NamespaceFilter() predicate.Funcs {
	funcs := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		if len(option.Opts.Namespaces) == 0 {
			return true
		}
		for _, ns := range option.Opts.Namespaces {
			if ns == obj.GetNamespace() {
				return true
			}
		}
		return false
	})

	return funcs
}

func getIngressBackends(ctx context.Context, r client.Reader, ingress *netv1.Ingress) (map[string]*corev1.Service, map[string]*corev1.Endpoints, error) {
	services := make(map[string]*corev1.Service)
	endpoints := make(map[string]*corev1.Endpoints)

	if len(option.Opts.DefaultBackend) > 0 {
		if svc, ep, err := getDefaultBackends(ctx, r, option.Opts.DefaultBackend); err == nil {
			services[option.Opts.DefaultBackend] = svc
			endpoints[option.Opts.DefaultBackend] = ep
		}
	}

	// if balance annotation exist, parse it to get service name
	balance, err := annotations.GetBalance(ingress.Annotations)
	if err != nil {
		return nil, nil, err
	}

	for _, rule := range ingress.Spec.Rules {
		for _, p := range rule.IngressRuleValue.HTTP.Paths {
			// service name exist in annotation
			var names []string
			if v, ok := balance[p.Backend.Service.Name]; ok {
				for name := range v {
					names = append(names, name)
				}
			} else {
				names = append(names, p.Backend.Service.Name)
			}

			for _, name := range names {
				if svc, err := getService(ctx, r, ingress.Namespace, name, p.Backend.Service.Port); err != nil {
					return nil, nil, err
				} else {
					services[util.NamespacedName(ingress.Namespace, name)] = svc
				}

				if ep, err := getEndpoint(ctx, r, ingress.Namespace, name); err != nil {
					return nil, nil, err
				} else {
					endpoints[util.NamespacedName(ingress.Namespace, name)] = ep
				}
			}

		}
	}

	return services, endpoints, nil
}

func getDefaultBackends(ctx context.Context, r client.Reader, name string) (*corev1.Service, *corev1.Endpoints, error) {
	// name is in format of "namespace/name"
	names := strings.Split(name, string(types.Separator))
	svc := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: names[0],
		Name:      names[1],
	}, svc)
	if err != nil {
		return nil, nil, err
	}

	if len(svc.Spec.Ports) == 0 {
		return nil, nil, fmt.Errorf("fail to get service: %s/%s", names[0], names[1])
	}

	ep, err := getEndpoint(ctx, r, names[0], names[1])
	if err != nil {
		return nil, nil, err
	}
	return svc, ep, nil
}

func getEndpoint(ctx context.Context, r client.Reader, namespace string, name string) (*corev1.Endpoints, error) {
	ep := &corev1.Endpoints{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, ep)
	if err != nil {
		return nil, err
	}
	if len(ep.Subsets) == 0 || len(ep.Subsets[0].Ports) == 0 {
		return nil, fmt.Errorf("not endpoint found for service, %s/%s", namespace, name)
	}
	return ep, nil
}

func getService(ctx context.Context, r client.Reader, namespace, name string, port netv1.ServiceBackendPort) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, svc)
	if err != nil {
		return nil, err
	}

	// check port exist
	for _, p := range svc.Spec.Ports {
		if p.Name == port.Name || p.Port == port.Number {
			return svc, nil
		}
	}

	return nil, fmt.Errorf("service[%s] port [%d %s] not found", name, port.Number, port.Name)
}

func getIngressSecret(ctx context.Context, r client.Reader, ingress *netv1.Ingress) ([]*corev1.Secret, error) {
	secrets := make([]*corev1.Secret, 0)
	for _, tls := range ingress.Spec.TLS {
		secret, err := getSecret(ctx, r, ingress.Namespace, tls.SecretName)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func getSecret(ctx context.Context, r client.Reader, namespace, name string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, secret)
	if err == nil {
		return secret, nil
	} else {
		return nil, err
	}
}

// set defaultBackend in ingress
func setDefautBackend(ingress *netv1.Ingress, service *corev1.Service) {
	if len(option.Opts.DefaultBackend) == 0 || service == nil || len(service.Spec.Ports) == 0 {
		return
	}

	names := strings.Split(option.Opts.DefaultBackend, "/")

	backend := &netv1.IngressBackend{}
	backend.Service = &netv1.IngressServiceBackend{
		Name: names[1],
		Port: netv1.ServiceBackendPort{
			Name:   service.Spec.Ports[0].Name,
			Number: service.Spec.Ports[0].Port},
	}

	ingress.Spec.DefaultBackend = backend
}
