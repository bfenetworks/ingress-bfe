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
	"sync"
	"syscall"
	"time"
)

import (
	"github.com/baidu/go-lib/log"
	"github.com/bfenetworks/bfe/bfe_util/signal_table"
	networking "k8s.io/api/networking/v1beta1"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/builder"
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

const (
	statusStaring = iota
	statusFailed
	statusStarted
)

type ingressList []*networking.Ingress

type BfeIngress struct {
	namespaces   []string
	labels       []string
	ingressClass string

	stopCh chan struct{}

	status int            // status for staring process
	wg     sync.WaitGroup // wait group for graceful exit

	BfeConfigRoot   string        // BFE config root path, for config dumping
	ReloadURLPrefix string        // common prefix for BFE reload URL
	SyncPeriod      time.Duration // period for ingress watcher to re-sync
}

func NewBfeIngress(namespaces, labels []string, ingressClass string) *BfeIngress {
	return &BfeIngress{
		namespaces:   namespaces,
		labels:       labels,
		ingressClass: ingressClass,
		stopCh:       make(chan struct{}),
		status:       statusStaring,
	}
}

func (ing *BfeIngress) Start() {
	ingressesCh := make(chan ingressList, 1)
	client, err := kubernetes_client.NewKubernetesClient()
	if err != nil {
		log.Logger.Warn("error in NewKubernetesClient(): %s", err)
	}

	ing.initSignalTable()
	ing.startWatcher(client, ingressesCh, ing.SyncPeriod)
	ing.startProcessor(client, ingressesCh)

	if ing.status == statusFailed {
		// exit if failed in somewhere
		ing.Shutdown(nil)
	} else {
		// update status as started
		ing.status = statusStarted
	}

	// waiting for shutdown
	ing.wg.Wait()
	log.Logger.Info("stop ingress")
}

// start ingress processor goroutine
func (ing *BfeIngress) startProcessor(client *kubernetes_client.KubernetesClient, ingressesCh chan ingressList) {
	// skip when ingress failed for some reason
	if ing.status == statusFailed {
		return
	}

	// create processor
	processor, err := NewProcessor(client, ingressesCh, ing.stopCh)
	if err != nil {
		log.Logger.Error(err)
		ing.status = statusFailed
		return
	}
	processor.dumper = builder.NewDumper(ing.BfeConfigRoot)
	processor.reloader = builder.NewReloader(ing.ReloadURLPrefix)

	// start processor goroutine
	ing.wg.Add(1)
	go processor.Start(&ing.wg)
}

// start ingress watcher goroutine
func (ing *BfeIngress) startWatcher(client *kubernetes_client.KubernetesClient, ingressesCh chan ingressList,
	syncPeriod time.Duration) {
	// skip when ingress failed for some reason
	if ing.status == statusFailed {
		return
	}

	// create watcher
	watcher, err := NewWatcher(ing.namespaces, ing.labels, ing.ingressClass, client, ingressesCh, ing.stopCh)
	if err != nil {
		log.Logger.Error(err)
		ing.status = statusFailed
		return
	}
	watcher.syncPeriod = syncPeriod

	// start watcher
	ing.wg.Add(1)
	go watcher.Start(&ing.wg)
}

// shutdown ingress by broadcasting stop signal
func (ing *BfeIngress) Shutdown(sig os.Signal) {
	close(ing.stopCh)
}

func (ing *BfeIngress) initSignalTable() {
	/* create signal table */
	signalTable := signal_table.NewSignalTable()

	/* register signal handlers */
	signalTable.Register(syscall.SIGQUIT, ing.Shutdown)
	signalTable.Register(syscall.SIGTERM, signal_table.TermHandler)
	signalTable.Register(syscall.SIGHUP, signal_table.IgnoreHandler)
	signalTable.Register(syscall.SIGILL, signal_table.IgnoreHandler)
	signalTable.Register(syscall.SIGTRAP, signal_table.IgnoreHandler)
	signalTable.Register(syscall.SIGABRT, signal_table.IgnoreHandler)

	/* start signal handler routine */
	signalTable.StartSignalHandle()
}
