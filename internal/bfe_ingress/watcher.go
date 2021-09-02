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

package bfe_ingress

import (
	"fmt"
	"github.com/bfenetworks/ingress-bfe/internal/utils"
	"reflect"
	"sort"
	"sync"
	"time"
)

import (
	"github.com/baidu/go-lib/log"
	"github.com/mitchellh/hashstructure"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

type IngressWatcher struct {
	namespace    []string
	labels       []string
	ingressClass string

	client     *kubernetes_client.KubernetesClient
	syncPeriod time.Duration // period for ingress watcher to re-sync
	ingressCh  chan ingressList
	stopCh     chan struct{}
}

var IngressService map[string]bool // ingress services watched

// NewWatcher creates watcher for ingress in k8s
func NewWatcher(namespaces []string, labels []string, ingressClass string,
	client *kubernetes_client.KubernetesClient, ingressCh chan ingressList,
	stopCh chan struct{}) (*IngressWatcher, error) {

	// check parameters
	if client == nil || ingressCh == nil || stopCh == nil {
		return nil, fmt.Errorf("create ingress watcher fail")
	}

	return &IngressWatcher{
		namespace:    namespaces,
		labels:       labels,
		ingressClass: ingressClass,

		client:     client,
		syncPeriod: utils.DefaultSyncPeriod,
		ingressCh:  ingressCh,
		stopCh:     stopCh,
	}, nil
}

func (iw *IngressWatcher) hash(ingressList []*networking.Ingress) (uint64, error) {
	cpIngressList := make([]*networking.Ingress, 0)
	for _, ingress := range ingressList {
		cpIngress := ingress.DeepCopy()
		if (*cpIngress).Annotations != nil {
			delete((*cpIngress).Annotations, StatusAnnotationKey)
			log.Logger.Info("name{%s} annotations{%v} spec.rules{%v}", (*cpIngress).Name, (*cpIngress).Annotations, (*cpIngress).Spec.Rules)
		}
		(*cpIngress).ObjectMeta.ResourceVersion = ""
		cpIngressList = append(cpIngressList, cpIngress)
	}
	sort.Slice(cpIngressList, func(i, j int) bool {
		if cpIngressList[i].CreationTimestamp.Equal(&cpIngressList[j].CreationTimestamp) {
			return cpIngressList[i].Name < cpIngressList[j].Name
		}
		return cpIngressList[i].CreationTimestamp.Before(&cpIngressList[j].CreationTimestamp)
	})
	return hashstructure.Hash(cpIngressList, nil)
}

func (iw *IngressWatcher) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	IngressService = make(map[string]bool)
	eventCh := iw.client.Watch(iw.namespace, iw.labels, iw.ingressClass, iw.syncPeriod)
	for {
		select {
		case msg := <-eventCh:
			t := reflect.TypeOf(msg).String()
			log.Logger.Debug("eventCh type is %s, eventCh message is %+v", t, msg)
			switch t {
			case "*v1beta1.Ingress":
				log.Logger.Info("process ingress resource")
				data := (msg).(*networking.Ingress)
				parseServiceFromIngress(data)
				log.Logger.Info("ingress services info: %v", IngressService)

			case "*v1.Endpoints":
				log.Logger.Info("process endpoints resource")
				data := (msg).(*core.Endpoints)
				endService := fmt.Sprintf("%s:%s", data.Namespace, data.Name)
				if _, ok := IngressService[endService]; !ok {
					continue
				}
			}

			ingresses := iw.client.GetIngresses()
			iw.ingressCh <- ingresses

		case <-iw.stopCh:
			log.Logger.Info("stop watcher")
			return
		}
	}
}

func parseServiceFromIngress(ingress *networking.Ingress) {
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			serviceName := path.Backend.ServiceName
			if serviceName != "" {
				IngressService[fmt.Sprintf("%s:%s", ingress.Namespace, serviceName)] = true
			}
		}
	}
}
