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

package extensionsv1beta1

import (
	"context"

	extv1beta1 "k8s.io/api/extensions/v1beta1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/filter"
	controllerV1 "github.com/bfenetworks/ingress-bfe/internal/controllers/netv1"
)

// IngressReconciler reconciles a extv1beta1 Ingress object
type IngressReconciler struct {
	BfeConfigBuilder *bfeConfig.ConfigBuilder

	client.Client
	Scheme *runtime.Scheme
}

func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("reconciling ingress", "api version", "ExtensionsV1beta1")

	// read ingress
	ingressExtV1beta1 := &extv1beta1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ingressExtV1beta1)
	if err != nil {
		r.BfeConfigBuilder.DeleteIngress(req.Namespace, req.Name)
		log.V(1).Info("reconcile: ingress delete")
		return reconcile.Result{}, err
	}

	if !filter.MatchIngressClass(ctx, r, ingressExtV1beta1.Annotations, ingressExtV1beta1.Spec.IngressClassName) {
		return reconcile.Result{}, nil
	}

	log.V(1).Info("reconcile: ingress object", "ingress", ingressExtV1beta1)

	ingressV1 := &netv1.Ingress{}
	convert(ingressExtV1beta1, ingressV1)

	err = controllerV1.ReconcileV1Ingress(ctx, r.Client, r.BfeConfigBuilder, ingressV1)
	setStatus(ctx, r.Client, err, ingressExtV1beta1)
	return reconcile.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extv1beta1.Ingress{}, builder.WithPredicates(filter.NamespaceFilter())).
		Complete(r)
}

func setStatus(ctx context.Context, r client.Client, err error, ingress *extv1beta1.Ingress) {
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

func convert(in *extv1beta1.Ingress, out *netv1.Ingress) {

	out.TypeMeta.Kind = "Ingress"
	out.TypeMeta.APIVersion = netv1.SchemeGroupVersion.String()

	out.ObjectMeta = *in.ObjectMeta.DeepCopy()
	out.ObjectMeta.ResourceVersion = "v1"

	if in.Spec.Backend != nil {
		out.Spec.DefaultBackend = &netv1.IngressBackend{
			Service: &netv1.IngressServiceBackend{
				Name: in.Spec.Backend.ServiceName,
				Port: netv1.ServiceBackendPort{
					Name:   in.Spec.Backend.ServicePort.StrVal,
					Number: in.Spec.Backend.ServicePort.IntVal,
				},
			},
			Resource: nil,
		}
	}

	out.Spec.IngressClassName = in.Spec.IngressClassName

	//Spec.TLS
	for _, tls := range in.Spec.TLS {
		outTLS := netv1.IngressTLS{
			Hosts:      tls.Hosts,
			SecretName: tls.SecretName,
		}
		out.Spec.TLS = append(out.Spec.TLS, outTLS)
	}

	for _, rule := range in.Spec.Rules {

		var paths []netv1.HTTPIngressPath

		for _, p := range rule.IngressRuleValue.HTTP.Paths {
			pathType := netv1.PathType(string(*p.PathType))
			path := netv1.HTTPIngressPath{
				Path:     p.Path,
				PathType: &pathType,
				Backend: netv1.IngressBackend{
					Service: &netv1.IngressServiceBackend{
						Name: p.Backend.ServiceName,
						Port: netv1.ServiceBackendPort{
							Name:   p.Backend.ServicePort.StrVal,
							Number: p.Backend.ServicePort.IntVal,
						},
					},
					Resource: nil,
				},
			}
			paths = append(paths, path)
		}

		ruleValue := &netv1.HTTPIngressRuleValue{
			Paths: paths,
		}
		outRule := netv1.IngressRule{
			Host:             rule.Host,
			IngressRuleValue: netv1.IngressRuleValue{HTTP: ruleValue},
		}
		out.Spec.Rules = append(out.Spec.Rules, outRule)
	}

}
