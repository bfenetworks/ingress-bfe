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

package kubernetes_client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/baidu/go-lib/log"

	corev1 "k8s.io/api/core/v1"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

const (
	AllIngressClass           = ""
	IngressClassAnnotationKey = "kubernetes.io/ingress.class"
)

type resourceEventHandler struct {
	ev chan<- interface{}
}

func (h *resourceEventHandler) OnAdd(obj interface{}) {
	eventHandlerFunc(h.ev, obj)
}

func (h *resourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	eventHandlerFunc(h.ev, newObj)
}

func (h *resourceEventHandler) OnDelete(obj interface{}) {
	eventHandlerFunc(h.ev, obj)
}

func eventHandlerFunc(events chan<- interface{}, obj interface{}) {
	select {
	case events <- obj:
	default:
	}
}

func newResourceEventHandler(events chan<- interface{}) cache.ResourceEventHandler {
	return &cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			return true
		},
		Handler: &resourceEventHandler{ev: events},
	}
}

type KubernetesCluster struct {
	SupportIngressClass bool
	SupportNetworking   bool
}

type KubernetesClient struct {
	namespaces []string
	watchAll   bool

	watchLabel bool
	labels     []string

	watchedIngressClass string

	clientset *kubernetes.Clientset
	factories map[string]informers.SharedInformerFactory
	eventCh   chan interface{}
	stopCh    chan struct{}

	cluster *KubernetesCluster
}

func NewKubernetesClient() (*KubernetesClient, error) {
	c := new(KubernetesClient)
	c.factories = make(map[string]informers.SharedInformerFactory)
	c.eventCh = make(chan interface{}, 1)
	c.stopCh = make(chan struct{})


	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("InClusterConfig error: %v", err)
	}
	c.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("NewForConfig error: %v", err)
	}

	c.cluster = &KubernetesCluster{
		SupportIngressClass: c.SupportIngressClassVersion(),
		SupportNetworking:   c.SupportIngressVersion(),
	}
	return c, nil
}

func (c *KubernetesClient) lookupNamespace(ns string) string {
	if c.watchAll {
		return metav1.NamespaceAll
	}
	return ns
}

func (c *KubernetesClient) Watch(namespaces []string, labels []string, ingressClass string, ResyncPeriod time.Duration) <-chan interface{} {
	c.namespaces = namespaces
	if len(namespaces) == 0 {
		c.namespaces = []string{metav1.NamespaceAll}
		c.watchAll = true
	}

	if c.watchAll && len(labels) > 0 {
		c.watchLabel = true
		c.labels = labels
	}

	c.watchedIngressClass = ingressClass

	eventHandler := newResourceEventHandler(c.eventCh)

	for _, ns := range c.namespaces {
		factory := informers.NewSharedInformerFactoryWithOptions(c.clientset, ResyncPeriod, informers.WithNamespace(ns))
		if c.cluster.SupportNetworking {
			factory.Networking().V1beta1().Ingresses().Informer().AddEventHandler(eventHandler)
		} else {
			factory.Extensions().V1beta1().Ingresses().Informer().AddEventHandler(eventHandler)
		}

		if c.cluster.SupportIngressClass {
			factory.Networking().V1beta1().IngressClasses().Informer().AddEventHandler(eventHandler)
		}

		factory.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		factory.Core().V1().Endpoints().Informer().AddEventHandler(eventHandler)
		factory.Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		if ns == metav1.NamespaceAll {
			factory.Core().V1().Namespaces().Informer().AddEventHandler(eventHandler)
		}

		go factory.Start(c.stopCh)
		c.factories[ns] = factory
	}

	return c.eventCh
}

func (c *KubernetesClient) Close() {
	close(c.stopCh)
}

func (c *KubernetesClient) GetEndpoints(namespace, name string) (*corev1.Endpoints, error) {
	endpoint, err := c.factories[c.lookupNamespace(namespace)].Core().V1().Endpoints().Lister().Endpoints(namespace).Get(name)
	return endpoint, err
}

func (c *KubernetesClient) GetService(namespace, name string) (*corev1.Service, error) {
	service, err := c.factories[c.lookupNamespace(namespace)].Core().V1().Services().Lister().Services(namespace).Get(name)
	return service, err
}

func (c *KubernetesClient) GetNamespaceByLabel() []*corev1.Namespace {
	if !c.watchAll || !c.watchLabel {
		return nil
	}
	factory := c.factories[metav1.NamespaceAll]
	labelsMap := make(map[string]string)
	for _, label := range c.labels {
		kV := strings.Split(label, "=")
		if len(kV) != 2 {
			continue
		}
		labelsMap[kV[0]] = kV[1]
	}
	labelSelector := labels.Set(labelsMap).AsSelector()
	namespaces, err := factory.Core().V1().Namespaces().Lister().List(labelSelector)
	if err != nil {
		log.Logger.Warn("fail to list namespace by label %s: %s", c.labels, err)
		return nil
	}
	return namespaces

}

func (c *KubernetesClient) GetSecretsByName(namespace, name string) (*corev1.Secret, error) {
	secret, err := c.factories[c.lookupNamespace(namespace)].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
	return secret, err
}

func (c *KubernetesClient) GetIngresses() []*networkingv1beta1.Ingress {
	var result []*networkingv1beta1.Ingress

	for ns, factory := range c.factories {
		ings := make([]*networkingv1beta1.Ingress, 0)
		var err error
		if !c.cluster.SupportNetworking {
			extendsIngs, err := factory.Extensions().V1beta1().Ingresses().Lister().List(labels.Everything())
			if err != nil {
				log.Logger.Info("Failed to list ingresses in namespace %s: %s", ns, err)
			}
			for _, ing := range extendsIngs {
				netIng, err := c.convertFromExtensions(ing)
				if err != nil {
					continue
				}
				ings = append(ings, netIng)
			}
		} else {
			ings, err = factory.Networking().V1beta1().Ingresses().Lister().List(labels.Everything())
			if err != nil {
				log.Logger.Info("Failed to list ingresses in namespace %s: %s", ns, err)
			}
		}
		if c.watchLabel {
			targetNamespaces := c.GetNamespaceByLabel()
			filterFunc := func(ing *networkingv1beta1.Ingress) bool {
				for _, targetNs := range targetNamespaces {
					if ing.Namespace == targetNs.GetName() {
						return true
					}
				}
				log.Logger.Debug("ns[%s] ingress[%s] filter by namespace", ing.Namespace, ing.Name)
				return false
			}
			ings = c.filterIngress(ings, filterFunc)
			result = append(result, ings...)
		} else {
			result = append(result, ings...)
		}
	}
	result = c.filterIngress(result, c.filterIngressByClass)
	return result
}

func (c *KubernetesClient) convertFromExtensions(old *extensionsv1beta1.Ingress) (*networkingv1beta1.Ingress, error) {
	data, err := old.Marshal()
	if err != nil {
		return nil, err
	}
	ni := &networkingv1beta1.Ingress{}
	err = ni.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return ni, nil
}

func (c *KubernetesClient) filterIngressByClass(ingress *networkingv1beta1.Ingress) bool {
	if c.watchedIngressClass == AllIngressClass {
		return true
	}
	if val, ok := ingress.Annotations[IngressClassAnnotationKey]; ok && val == c.watchedIngressClass {
		return true
	}

	if c.cluster.SupportIngressClass && ingress.Spec.IngressClassName != nil {
		ns := c.lookupNamespace(ingress.Namespace)
		ic, err := c.factories[ns].Networking().V1beta1().IngressClasses().Lister().Get(*ingress.Spec.IngressClassName)

		if err != nil || ic == nil {
			return false
		}
		return true
	}
	log.Logger.Debug("ns[%s] ingress[%s] filter by ingress.class[%s]", ingress.Namespace,
		ingress.Name, c.watchedIngressClass)
	return false
}

func (c *KubernetesClient) GetVersion() (*version.Info, error) {
	return c.clientset.Discovery().ServerVersion()
}

func (c *KubernetesClient) SupportIngressClassVersion() bool {
	serverVersion, err := c.GetVersion()
	log.Logger.Info("get server running version %v", serverVersion)
	if err != nil {
		return false
	}
	v118, _ := versionutil.ParseGeneric("v1.18.0")
	runningVersion, err := versionutil.ParseGeneric(serverVersion.String())
	if err != nil {
		return false
	}
	return runningVersion.AtLeast(v118)
}

func (c *KubernetesClient) SupportIngressVersion() bool {
	serverVersion, err := c.GetVersion()
	log.Logger.Info("get server running version %v", serverVersion)
	if err != nil {
		return false
	}
	v114, _ := versionutil.ParseGeneric("v1.14.0")
	runningVersion, err := versionutil.ParseGeneric(serverVersion.String())
	if err != nil {
		return false
	}
	return runningVersion.AtLeast(v114)
}

func (c *KubernetesClient) UpdateIngressAnnotation(namespace, name, annotation, msg string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ingressClient := c.clientset.ExtensionsV1beta1().Ingresses(namespace)
		result, getErr := ingressClient.Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of Ingress: %v", getErr)
		}
		if result.Annotations == nil {
			result.Annotations = make(map[string]string)
		}
		if val, exists := result.Annotations[annotation]; exists {
			if val == msg {
				return nil
			}
		}
		result.Annotations[annotation] = msg
		_, updateErr := ingressClient.Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})
	return retryErr
}

func (c *KubernetesClient) filterIngress(ingresses []*networkingv1beta1.Ingress, filter IngressFilterFunc) []*networkingv1beta1.Ingress {
	result := make([]*networkingv1beta1.Ingress, 0)
	for _, ing := range ingresses {
		if ing != nil && filter(ing) {
			result = append(result, ing)
		}
	}
	return result
}

type IngressFilterFunc func(*networkingv1beta1.Ingress) bool
