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

package bfeConfig

import (
	"context"
	"sync"
	"time"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/modules"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs"
	"github.com/bfenetworks/ingress-bfe/internal/option"
)

var (
	log = ctrl.Log.WithName("configBuilder")
)

type ConfigBuilder struct {
	lock sync.Mutex

	serverDataConf *configs.ServerDataConfig
	clusterConf    *configs.ClusterConfig
	tlsConf        *configs.TLSConfig
	modules        []modules.BFEModuleConfig
}

func NewConfigBuilder() *ConfigBuilder {
	version := "init"
	return &ConfigBuilder{
		serverDataConf: configs.NewServerDataConfig(version),
		clusterConf:    configs.NewClusterConfig(version),
		tlsConf:        configs.NewTLSConfig(version),
		modules:        modules.InitBFEModules(version),
	}
}

func (c *ConfigBuilder) UpdateIngress(ingress *netv1.Ingress, services map[string]*corev1.Service, endpoints map[string]*corev1.Endpoints, secrets []*corev1.Secret) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.serverDataConf.UpdateIngress(ingress); err != nil {
		return err
	}

	if err := c.clusterConf.UpdateIngress(ingress, services, endpoints); err != nil {
		c.serverDataConf.DeleteIngress(ingress.Namespace, ingress.Name)
		return err
	}

	// update secret
	if err := c.tlsConf.UpdateIngress(ingress, secrets); err != nil {
		c.clusterConf.DeleteIngress(ingress.Namespace, ingress.Name)
		c.serverDataConf.DeleteIngress(ingress.Namespace, ingress.Name)
		return err
	}

	// update modules
	for i, module := range c.modules {
		if err := module.UpdateIngress(ingress); err != nil {
			c.clusterConf.DeleteIngress(ingress.Namespace, ingress.Name)
			c.serverDataConf.DeleteIngress(ingress.Namespace, ingress.Name)
			c.tlsConf.DeleteIngress(ingress.Namespace, ingress.Name)
			for j := 0; j < i; j++ {
				c.modules[j].DeleteIngress(ingress.Namespace, ingress.Name)
			}
			return err
		}
	}

	return nil
}

func (c *ConfigBuilder) DeleteIngress(namespace, name string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.serverDataConf.DeleteIngress(namespace, name)
	c.clusterConf.DeleteIngress(namespace, name)
	c.tlsConf.DeleteIngress(namespace, name)

	for _, module := range c.modules {
		module.DeleteIngress(namespace, name)
	}
}

func (c *ConfigBuilder) UpdateService(service *corev1.Service, endpoint *corev1.Endpoints) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_ = c.clusterConf.UpdateService(service, endpoint)
}

func (c *ConfigBuilder) DeleteService(namespace, name string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.clusterConf.DeleteService(namespace, name)
}

func (c *ConfigBuilder) UpdateSecret(secret *corev1.Secret) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.tlsConf.UpdateSecret(secret); err != nil {
		return err
	}
	return nil
}

func (c *ConfigBuilder) DeleteSecret(namespace, name string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tlsConf.DeleteSecret(namespace, name)
}

func (c *ConfigBuilder) InitReload(ctx context.Context) {
	tick := time.NewTicker(option.Opts.Ingress.ReloadInterval)

	go func() {
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				if err := c.reload(); err != nil {
					log.Error(err, "fail to reload config")
				}
			case <-ctx.Done():
				log.Info("exit bfe reload")
				return
			}
		}
	}()

}

func (c *ConfigBuilder) reload() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.serverDataConf.Reload(); err != nil {
		log.Error(err, "Fail to reload config",
			"serverDataConf",
			c.serverDataConf)
		return err
	}

	if err := c.clusterConf.Reload(); err != nil {
		log.Error(err, "Fail to reload config",
			"clusterConf",
			c.clusterConf)
		return err
	}

	if err := c.tlsConf.Reload(); err != nil {
		log.Error(err, "Fail to reload config",
			"tlsConf",
			c.tlsConf)
		return err
	}

	for _, module := range c.modules {
		if err := module.Reload(); err != nil {
			log.Error(err, "Fail to reload config",
				module.Name(),
				module)
			return err
		}
	}

	return nil
}
