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
	"reflect"
	"strings"
	"time"
)

import (
	"github.com/baidu/go-lib/log"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	util "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/core/v1"
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
	eventHandlerFunc(h.ev, obj, "add")
}

func (h *resourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}
	log.Logger.Debug("oldObj{%v} newObj{%v}", oldObj, newObj)
	eventHandlerFunc(h.ev, newObj, "update")
}

func (h *resourceEventHandler) OnDelete(obj interface{}) {
	eventHandlerFunc(h.ev, obj, "del")
}

func eventHandlerFunc(events chan<- interface{}, obj interface{}, action string) {
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
		return meta.NamespaceAll
	}
	return ns
}

func (c *KubernetesClient) Watch(namespaces []string, labels []string, ingressClass string, ResyncPeriod time.Duration) <-chan interface{} {
	c.namespaces = namespaces
	if len(namespaces) == 0 {
		c.namespaces = []string{meta.NamespaceAll}
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

		resources := factory.Core().V1()
		resources.Services().Informer().AddEventHandler(eventHandler)
		resources.Endpoints().Informer().AddEventHandler(eventHandler)
		resources.Secrets().Informer().AddEventHandler(eventHandler)
		if ns == meta.NamespaceAll {
			resources.Namespaces().Informer().AddEventHandler(eventHandler)
		}

		go factory.Start(c.stopCh)
		c.factories[ns] = factory
	}

	return c.eventCh
}

func (c *KubernetesClient) Close() {
	close(c.stopCh)
}

func (c *KubernetesClient) GetResources(namespace string) v1.Interface {
	return c.factories[c.lookupNamespace(namespace)].Core().V1()
}

func (c *KubernetesClient) GetEndpoints(namespace, name string) (*core.Endpoints, error) {
	endpoint, err := c.GetResources(namespace).Endpoints().Lister().Endpoints(namespace).Get(name)
	return endpoint, err
}

func (c *KubernetesClient) GetService(namespace, name string) (*core.Service, error) {
	service, err := c.GetResources(namespace).Services().Lister().Services(namespace).Get(name)
	return service, err
}

func (c *KubernetesClient) GetNamespaceByLabel() []*core.Namespace {
	if !c.watchAll || !c.watchLabel {
		return nil
	}
	labelsMap := make(map[string]string)
	for _, label := range c.labels {
		kV := strings.Split(label, "=")
		if len(kV) != 2 {
			continue
		}
		labelsMap[kV[0]] = kV[1]
	}
	labelSelector := labels.Set(labelsMap).AsSelector()
	namespaces, err := c.GetResources(meta.NamespaceAll).Namespaces().Lister().List(labelSelector)
	if err != nil {
		log.Logger.Warn("fail to list namespace by label %s: %s", c.labels, err)
		return nil
	}
	return namespaces
}

func (c *KubernetesClient) GetSecretsByName(namespace, name string) (*core.Secret, error) {
	secret, err := c.GetResources(namespace).Secrets().Lister().Secrets(namespace).Get(name)
	return secret, err
}

func (c *KubernetesClient) GetIngresses() []*networking.Ingress {
	var result []*networking.Ingress

	for ns, factory := range c.factories {
		ings, err := c.getAllIngresses(factory)
		if err != nil {
			log.Logger.Info("Failed to list ingresses in namespace %s: %s", ns, err)
			continue
		}

		if c.watchLabel {
			targetNamespaces := c.GetNamespaceByLabel()
			filterFunc := func(ing *networking.Ingress) bool {
				for _, targetNs := range targetNamespaces {
					if ing.Namespace == targetNs.GetName() {
						return true
					}
				}
				log.Logger.Debug("ns[%s] ingress[%s] filter by namespace", ing.Namespace, ing.Name)
				return false
			}
			ings = c.filterIngress(ings, filterFunc)
		}
		result = append(result, ings...)
	}
	return c.filterIngress(result, c.filterIngressByClass)
}

// getAllIngresses gets all ingresses from certain informer factory
func (c *KubernetesClient) getAllIngresses(factory informers.SharedInformerFactory) ([]*networking.Ingress, error) {
	if !c.cluster.SupportNetworking {
		extendsIngs, err := factory.Extensions().V1beta1().Ingresses().Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}

		ings := make([]*networking.Ingress, 0)
		for _, ing := range extendsIngs {
			netIng, err := c.convertFromExtensions(ing)
			if err != nil {
				continue
			}
			ings = append(ings, netIng)
		}
		return ings, nil
	}

	return factory.Networking().V1beta1().Ingresses().Lister().List(labels.Everything())
}

func (c *KubernetesClient) convertFromExtensions(old *extensions.Ingress) (*networking.Ingress, error) {
	data, err := old.Marshal()
	if err != nil {
		return nil, err
	}
	ni := &networking.Ingress{}
	err = ni.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return ni, nil
}

func (c *KubernetesClient) filterIngressByClass(ingress *networking.Ingress) bool {
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
	v118, _ := util.ParseGeneric("v1.18.0")
	runningVersion, err := util.ParseGeneric(serverVersion.String())
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
	v114, _ := util.ParseGeneric("v1.14.0")
	runningVersion, err := util.ParseGeneric(serverVersion.String())
	if err != nil {
		return false
	}
	return runningVersion.AtLeast(v114)
}

func (c *KubernetesClient) UpdateIngressAnnotation(namespace, name, annotation, msg string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ingressClient := c.clientset.ExtensionsV1beta1().Ingresses(namespace)
		result, getErr := ingressClient.Get(context.TODO(), name, meta.GetOptions{})
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
		_, updateErr := ingressClient.Update(context.TODO(), result, meta.UpdateOptions{})
		return updateErr
	})
	return retryErr
}

func (c *KubernetesClient) filterIngress(ingresses []*networking.Ingress, filter IngressFilterFunc) []*networking.Ingress {
	result := make([]*networking.Ingress, 0)
	for _, ing := range ingresses {
		if ing != nil && filter(ing) {
			result = append(result, ing)
		}
	}
	return result
}

type IngressFilterFunc func(*networking.Ingress) bool
