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
	"sort"
	"sync"
	"time"
)

import (
	"github.com/baidu/go-lib/log"
	networking "k8s.io/api/networking/v1beta1"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/builder"
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
	"github.com/bfenetworks/ingress-bfe/internal/utils"
)

type Processor struct {
	client    *kubernetes_client.KubernetesClient
	ingressCh chan ingressList
	stopCh    chan struct{}
	reloader  *builder.Reloader

	statusWriter *IngressStatusWriter
}

func (p *Processor) initBfeConfigBuilder() []builder.BfeConfigBuilder {
	version := "reload"
	var builders []builder.BfeConfigBuilder

	builders = append(builders, builder.NewBfeBalanceConfigBuilder(p.client, version, p.reloader))
	builders = append(builders, builder.NewBfeRouteConfigBuilder(p.client, version, p.reloader))
	builders = append(builders, builder.NewBfeTLSConfigBuilder(p.client, version, p.reloader))

	return builders
}

func (p *Processor) processIngresses(ingresses ingressList) {
	cur := time.Now().UTC().String()

	builders := p.initBfeConfigBuilder()

	for _, ingress := range ingresses {
		log.Logger.Info("time[%s] ingress: namespaces[%s], ingress[%s], stamp[%s]", cur, ingress.Namespace, ingress.Name, ingress.CreationTimestamp.Time.String())

		var submittedBuilders = make([]builder.BfeConfigBuilder, 0)
		unSubmitted := false

		// submit current ingress to different type of config builders
		for _, builder := range builders {
			err := builder.Submit(ingress)
			if err != nil {
				log.Logger.Warn("namespaces[%s] ingress[%s] submit error[%s]", ingress.Namespace, ingress.Name, err.Error())
				unSubmitted = true
				p.doRollback(submittedBuilders, ingress)
				p.setStatus(ingress, true, err.Error())
				break
			} else {
				submittedBuilders = append(submittedBuilders, builder)
			}
		}
		if !unSubmitted {
			p.setStatus(ingress, false, "")
		}
	}
	if err := p.build(builders); err != nil {
		return
	}
	if err := p.dump(builders); err != nil {
		return
	}
	if err := p.reload(builders); err != nil {
		return
	}
}

func (p *Processor) build(builders []builder.BfeConfigBuilder) error {
	for _, builder := range builders {
		err := builder.Build()
		if err != nil {
			log.Logger.Warn("builder build error[%s]", err.Error())
			return err
		}
	}
	return nil
}

func (p *Processor) dump(builders []builder.BfeConfigBuilder) error {
	for _, builder := range builders {
		err := builder.Dump()
		if err != nil {
			log.Logger.Warn("builder dump error[%s]", err.Error())
			return err
		}
	}
	return nil
}

func (p *Processor) reload(builders []builder.BfeConfigBuilder) error {
	for _, builder := range builders {
		err := builder.Reload()
		if err != nil {
			log.Logger.Warn("builder reload error[%s]", err.Error())
			return err
		}
	}
	return nil
}

// when route config conflict, the older config will win, so sort ingress by create time
func (p *Processor) sortIngresses(ingresses ingressList) {
	sort.Slice(ingresses, func(i, j int) bool {
		if ingresses[i].CreationTimestamp.Equal(&ingresses[j].CreationTimestamp) {
			return ingresses[i].Name < ingresses[j].Name
		}
		return ingresses[i].CreationTimestamp.Before(&ingresses[j].CreationTimestamp)
	})
}

func (p *Processor) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case ingresses := <-p.ingressCh:
			log.Logger.Info("process [%d] ingress", len(ingresses))
			p.sortIngresses(ingresses)
			p.processIngresses(ingresses)
		case <-p.stopCh:
			log.Logger.Info("stop processor")
			return
		}
	}
}

func NewProcessor(c *kubernetes_client.KubernetesClient, ingressCh chan ingressList,
	stopCh chan struct{}) (*Processor, error) {

	// check parameters
	if c == nil || ingressCh == nil || stopCh == nil {
		return nil, fmt.Errorf("create processor fail")
	}

	return &Processor{
		client:       c,
		ingressCh:    ingressCh,
		stopCh:       stopCh,
		reloader:     builder.NewReloader(utils.ReloadUrlPrefix),
		statusWriter: &IngressStatusWriter{client: c},
	}, nil
}

func (p *Processor) doRollback(builders []builder.BfeConfigBuilder, ingress *networking.Ingress) {
	for _, builder := range builders {
		log.Logger.Debug("rollback namespaces[%s] ingress[%s]", ingress.Namespace, ingress.Name)
		err := builder.Rollback(ingress)
		if err != nil {
			log.Logger.Warn("namespaces[%s] ingress[%s] submit error[%s]", ingress.Namespace, ingress.Name, err.Error())
		}
	}
}

func (p *Processor) setStatus(ingress *networking.Ingress, err bool, msg string) {
	if err {
		p.statusWriter.SetError(ingress.Namespace, ingress.Name, msg)
	} else {
		p.statusWriter.SetSuccess(ingress.Namespace, ingress.Name)
	}
}
