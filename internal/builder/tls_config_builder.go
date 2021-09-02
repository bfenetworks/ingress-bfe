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
package builder

import (
	"fmt"
	"reflect"
)

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_tls_conf/server_cert_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_tls_conf/tls_rule_conf"
	networking "k8s.io/api/networking/v1beta1"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

const (
	DefaultCNName  = "example.org"
	DefaultCrtPath = "tls_conf/certs/example.crt"
	DefaultCrtKey  = "tls_conf/certs/example.key"
)

type certKeyConf struct {
	cert []byte
	key  []byte
}

type BfeTLSConf struct {
	serverCertConf server_cert_conf.BfeServerCertConf
	certKeyConf    map[string]certKeyConf
	tlsRuleConf    tls_rule_conf.BfeTlsRuleConf
}

var (
	ServerCertData    = "tls_conf/server_cert_conf.data"
	TLSRuleData       = "tls_conf/tls_rule_conf.data"
	CertKeyFilePath   = "tls_conf/certs/"
	ConfigNameTLSConf = "tls_conf"

	SecretCrt = "tls.crt"
	SecretKey = "tls.key"
)

type BfeTLSConfigBuilder struct {
	client *kubernetes_client.KubernetesClient

	dumper   *Dumper
	reloader *Reloader

	version string

	serverCertConf server_cert_conf.BfeServerCertConf

	certKeyConf  map[string]certKeyConf
	hostRefCount map[string]int

	tc *BfeTLSConf
}

func NewBfeTLSConfigBuilder(client *kubernetes_client.KubernetesClient, version string, dumper *Dumper,
	reloader *Reloader) *BfeTLSConfigBuilder {
	c := &BfeTLSConfigBuilder{
		client:       client,
		dumper:       dumper,
		reloader:     reloader,
		version:      version,
		certKeyConf:  make(map[string]certKeyConf),
		hostRefCount: make(map[string]int),
	}
	return c
}

func (c *BfeTLSConfigBuilder) CheckTLS(crt, pkey []byte, host string) bool {
	// TODO: add pem check in this function
	return true
}

func (c *BfeTLSConfigBuilder) CheckTlsConflict(certKeyConf map[string]certKeyConf) bool {
	for host, certKey := range certKeyConf {
		if _, ok := c.certKeyConf[host]; ok {
			if !reflect.DeepEqual(c.certKeyConf[host], certKey) {
				return false
			}
		}
	}
	return true
}

func (c *BfeTLSConfigBuilder) submitCertKeyMap(certKeyMap map[string]certKeyConf) error {
	if !c.CheckTlsConflict(certKeyMap) {
		var keys []string
		for key := range certKeyMap {
			keys = append(keys, key)
		}
		return fmt.Errorf("cert conflict in host %v", keys)
	}
	for host, cert := range certKeyMap {
		c.certKeyConf[host] = cert
		if _, ok := c.hostRefCount[host]; !ok {
			c.hostRefCount[host] = 0
		}
		c.hostRefCount[host]++
	}
	return nil
}

func (c *BfeTLSConfigBuilder) getCertKeyMap(ingress *networking.Ingress) (map[string]certKeyConf, error) {
	certKeyMap := make(map[string]certKeyConf)
	namespace := ingress.Namespace
	for _, tlsRule := range ingress.Spec.TLS {
		secretName := tlsRule.SecretName
		secrets, err := c.client.GetSecretsByName(namespace, secretName)
		if err != nil {
			return nil, fmt.Errorf("submit ingress %s fail, get secrets err: %s", ingress.Name, err.Error())
		}
		if _, exists := secrets.Data[SecretKey]; !exists {
			return nil, fmt.Errorf("submit ingress %s tls error: %s secret has no %s", ingress.Name, secretName, SecretKey)
		}
		if _, exists := secrets.Data[SecretCrt]; !exists {
			return nil, fmt.Errorf("submit ingress %s tls error: %s secret has no %s", ingress.Name, secretName, SecretCrt)
		}
		var crt = secrets.Data[SecretCrt]
		var key = secrets.Data[SecretKey]

		Hosts := tlsRule.Hosts
		for _, host := range Hosts {
			if !c.CheckTLS(crt, key, host) {
				return nil, fmt.Errorf("submit ingress tls error: check %s for host %s crt/key error ", secretName, host)
			}
			certKeyMap[host] = certKeyConf{
				cert: crt,
				key:  key,
			}
		}
	}
	return certKeyMap, nil
}

func (c *BfeTLSConfigBuilder) Submit(ingress *networking.Ingress) error {
	certKeyMap, err := c.getCertKeyMap(ingress)
	if err != nil {
		return err
	}
	return c.submitCertKeyMap(certKeyMap)
}

func (c *BfeTLSConfigBuilder) Rollback(ingress *networking.Ingress) error {
	for _, tlsRule := range ingress.Spec.TLS {
		Hosts := tlsRule.Hosts
		for _, host := range Hosts {
			if _, ok := c.hostRefCount[host]; ok {
				c.hostRefCount[host]--
				if c.hostRefCount[host] <= 0 {
					delete(c.hostRefCount, host)
					delete(c.certKeyConf, host)
				}
			}
		}
	}
	return nil
}

func (c *BfeTLSConfigBuilder) Build() error {
	if len(c.certKeyConf) == 0 {
		return c.buildDefault()
	}
	return c.buildCustom()
}

func (c *BfeTLSConfigBuilder) buildDefault() error {
	c.tc = &BfeTLSConf{
		serverCertConf: server_cert_conf.BfeServerCertConf{
			Version: c.version,
			Config: server_cert_conf.ServerCertConfMap{
				Default: DefaultCNName,
				CertConf: map[string]server_cert_conf.ServerCertConf{
					DefaultCNName: {
						ServerCertFile: DefaultCrtPath,
						ServerKeyFile:  DefaultCrtKey,
					},
				},
			},
		},
	}
	c.buildTLSConfig()
	return nil
}

func (c *BfeTLSConfigBuilder) buildTLSConfig() error {
	c.tc.tlsRuleConf = tls_rule_conf.BfeTlsRuleConf{
		Version:              c.version,
		Config:               map[string]*tls_rule_conf.TlsRuleConf{},
		DefaultChacha20:      false,
		DefaultDynamicRecord: false,
		DefaultNextProtos:    []string{"http/1.1"},
	}
	return nil
}

func (c *BfeTLSConfigBuilder) buildCustom() error {
	c.tc = new(BfeTLSConf)
	var sc server_cert_conf.BfeServerCertConf

	sc.Version = c.version
	sc.Config.CertConf = make(map[string]server_cert_conf.ServerCertConf)

	c.tc.certKeyConf = c.certKeyConf
	c.tc.serverCertConf = sc
	defaultHost := ""
	for host := range c.certKeyConf {
		if defaultHost == "" {
			defaultHost = host
		}
		if defaultHost > host {
			defaultHost = host
		}
		sc.Config.CertConf[host] = server_cert_conf.ServerCertConf{
			ServerCertFile: c.getCertFilePath(host),
			ServerKeyFile:  c.getKeyFilePath(host),
		}
	}
	c.tc.serverCertConf.Config.Default = defaultHost
	c.buildTLSConfig()
	return nil
}

func (c *BfeTLSConfigBuilder) Dump() error {
	// dump key and cert for hosts
	for host, ck := range c.tc.certKeyConf {
		certFile := c.getCertFilePath(host)
		if err := c.dumper.DumpBytes(ck.cert, certFile); err != nil {
			return fmt.Errorf("write [%s] cert file fail, err: %s", host, err)
		}

		keyFile := c.getKeyFilePath(host)
		if err := c.dumper.DumpBytes(ck.key, keyFile); err != nil {
			return fmt.Errorf("write [%s] key file fail, err: %s", host, err)
		}
	}

	// dump server cert config
	err := c.dumper.DumpJson(c.tc.serverCertConf, ServerCertData)
	if err != nil {
		return fmt.Errorf("dump server_cert_conf: %v", err)
	}

	// dump TLS rule config
	err = c.dumper.DumpJson(c.tc.tlsRuleConf, TLSRuleData)
	if err != nil {
		return fmt.Errorf("dump tls_rule_conf: %v", err)
	}

	return nil
}

func (c *BfeTLSConfigBuilder) Reload() error {
	return c.reloader.DoReload(c.tc, ConfigNameTLSConf)
}

func (c *BfeTLSConfigBuilder) getCertFilePath(host string) string {
	return c.dumper.Join(CertKeyFilePath, host+".cer")
}

func (c *BfeTLSConfigBuilder) getKeyFilePath(host string) string {
	return c.dumper.Join(CertKeyFilePath, host+".key")
}
