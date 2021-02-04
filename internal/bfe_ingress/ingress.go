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
	"os"
	"sort"
	"time"
)
import (
	"github.com/baidu/go-lib/log"
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

var (
	ResyncPeriod = 20 * time.Second
	ConfigPath   = "/bfe/output/conf/"
)

const (
	FilePerm os.FileMode = 0744
)

func WatchIngress(namespaces []string, labels []string, ingressClass string) {
	k8sClient, err := kubernetes_client.NewKubernetesClient()
	if err != nil {
		log.Logger.Error("NewK8sClient failed: %v", err)
		return
	}

	lastConfigMap = make(map[string]interface{})

	eventCh := k8sClient.Watch(namespaces, labels, ingressClass, ResyncPeriod)

	for {
		select {
		case <-eventCh:
			ingresses := k8sClient.GetIngresses()
			// when route config conflict, the older config will win, so sort ingress by create time
			sort.Slice(ingresses, func(i, j int) bool {
				if ingresses[i].CreationTimestamp.Equal(&ingresses[j].CreationTimestamp) {
					return ingresses[i].Name < ingresses[j].Name
				}
				return ingresses[i].CreationTimestamp.Before(&ingresses[j].CreationTimestamp)
			})
			doReload(k8sClient, ingresses)
		}
	}
}

func doReload(c *kubernetes_client.KubernetesClient, ingresses []*networkingv1beta1.Ingress) {
	cur := time.Now().UTC().String()
	version := "reload"

	var ingressConfigs []BfeIngressConfig

	ingressConfigs = append(ingressConfigs, NewBfeBalanceIngressConfig(c, version))
	ingressConfigs = append(ingressConfigs, NewBfeRouteIngressConfig(c, version))
	ingressConfigs = append(ingressConfigs, NewBfeTlsIngressConfig(c, version))

	for _, ingress := range ingresses {
		log.Logger.Info("time[%s] ingress: namespace[%s], ingress[%s], stamp[%s]", cur, ingress.Namespace, ingress.Name, ingress.CreationTimestamp.Time.String())
		var submitedConfigs = make([]BfeIngressConfig, 0)
		isRollbacked := false
		for _, ingressConfig := range ingressConfigs {
			err := ingressConfig.Submit(ingress)
			if err != nil {
				log.Logger.Warn("namespace[%s] ingress[%s] submit error[%s]", ingress.Namespace, ingress.Name, err.Error())
				isRollbacked = true
				doRollback(submitedConfigs, ingress)
				setStatus(c, ingress, true, err.Error())
				break
			} else {
				submitedConfigs = append(submitedConfigs, ingressConfig)
			}
		}
		if !isRollbacked {
			setStatus(c, ingress, false, "")
		}
	}
	for _, ingressConfig := range ingressConfigs {
		err := ingressConfig.Build()
		if err != nil {
			log.Logger.Warn("ingressConfig build error[%s]", err.Error())
			return
		}
	}
	for _, ingressConfig := range ingressConfigs {
		err := ingressConfig.Dump()
		if err != nil {
			log.Logger.Warn("ingressConfig dump error[%s]", err.Error())
			return
		}
	}
	for _, ingressConfig := range ingressConfigs {
		err := ingressConfig.Reload()
		if err != nil {
			log.Logger.Warn("ingressConfig reload error[%s]", err.Error())
			return
		}
	}
}

func doRollback(configList []BfeIngressConfig, ingress *networkingv1beta1.Ingress) {
	for _, config := range configList {
		log.Logger.Debug("rollback namespace[%s] ingress[%s]", ingress.Namespace, ingress.Name)
		err := config.Rollback(ingress)
		if err != nil {
			log.Logger.Warn("namespace[%s] ingress[%s] submit error[%s]", ingress.Namespace, ingress.Name, err.Error())
		}
	}
}

func setStatus(c *kubernetes_client.KubernetesClient, ingress *networkingv1beta1.Ingress, err bool, msg string) {
	var ingressStatus = IngressStatusWriter{
		client: c,
	}
	if err {
		ingressStatus.SetError(ingress.Namespace, ingress.Name, msg)
	} else {
		ingressStatus.SetSuccess(ingress.Namespace, ingress.Name)
	}
}

type BfeIngressConfig interface {
	Submit(ingress *networkingv1beta1.Ingress) error
	Rollback(ingress *networkingv1beta1.Ingress) error
	Build() error
	Dump() error
	Reload() error
}
