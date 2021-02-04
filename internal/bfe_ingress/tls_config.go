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
	"io/ioutil"
	"reflect"

	"github.com/bfenetworks/bfe/bfe_config/bfe_tls_conf/server_cert_conf"
	"github.com/bfenetworks/bfe/bfe_util"
	"github.com/baidu/go-lib/log"

	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
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
}

var (
	ServerCertData    = "tls_conf/server_cert_conf.data"
	CertKeyFilePath   = "tls_conf/certs/"
	ConfigNameTLSConf = "tls_conf"

	SecretCrt = "tls.crt"
	SecretKey = "tls.key"
)

type BfeTLSIngressConfig struct {
	client  *kubernetes_client.KubernetesClient
	version string

	serverCertConf server_cert_conf.BfeServerCertConf

	certKeyConf  map[string]certKeyConf
	hostRefCount map[string]int

	tc *BfeTLSConf
}

func NewBfeTlsIngressConfig(client *kubernetes_client.KubernetesClient, version string) *BfeTLSIngressConfig {
	tlsConfig := &BfeTLSIngressConfig{}
	tlsConfig.client = client
	tlsConfig.version = version
	tlsConfig.certKeyConf = make(map[string]certKeyConf)
	tlsConfig.hostRefCount = make(map[string]int)
	return tlsConfig
}

func (c *BfeTLSIngressConfig) CheckTLS(crt, pkey []byte, host string) bool {
	// TODO: add pem check in this function
	return true
}

func (c *BfeTLSIngressConfig) CheckTlsConflict(certKeyConf map[string]certKeyConf) bool {
	for host, certKey := range certKeyConf {
		if _, ok := c.certKeyConf[host]; ok {
			if !reflect.DeepEqual(c.certKeyConf[host], certKey) {
				return false
			}
		}
	}
	return true
}

func (c *BfeTLSIngressConfig) submitCertKeyMap(certKeyMap map[string]certKeyConf) error {
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
func (c *BfeTLSIngressConfig) getCertKeyMap(ingress *networkingv1beta1.Ingress) (map[string]certKeyConf, error) {
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

func (c *BfeTLSIngressConfig) Submit(ingress *networkingv1beta1.Ingress) error {
	certKeyMap, err := c.getCertKeyMap(ingress)
	if err != nil {
		return err
	}
	return c.submitCertKeyMap(certKeyMap)
}

func (c *BfeTLSIngressConfig) Rollback(ingress *networkingv1beta1.Ingress) error {
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

func (c *BfeTLSIngressConfig) Build() error {
	if len(c.certKeyConf) == 0 {
		return c.buildDefault()
	}
	return c.buildCustom()
}

func (c *BfeTLSIngressConfig) buildDefault() error {
	c.tc = new(BfeTLSConf)
	var sc server_cert_conf.BfeServerCertConf
	sc.Config.Default = DefaultCNName
	sc.Config.CertConf = make(map[string]server_cert_conf.ServerCertConf)
	sc.Config.CertConf[DefaultCNName] = server_cert_conf.ServerCertConf{
		ServerCertFile: DefaultCrtPath,
		ServerKeyFile:  DefaultCrtKey,
	}
	c.tc.serverCertConf = sc
	return nil
}

func (c *BfeTLSIngressConfig) buildCustom() error {
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
	return nil
}

func (c *BfeTLSIngressConfig) Dump() error {
	for host, ck := range c.tc.certKeyConf {
		certFile := c.getCertFilePath(host)
		keyFile := c.getKeyFilePath(host)
		err := ioutil.WriteFile(certFile, ck.cert, FilePerm)
		if err != nil {
			log.Logger.Info("write cert file fail, err: %s", err)
		}

		err = ioutil.WriteFile(keyFile, ck.key, FilePerm)
		if err != nil {
			log.Logger.Info("write cert file fail, err: %s", err)
		}
	}
	certConfFile := c.getcertConfFilePath()
	err := bfe_util.DumpJson(c.tc.serverCertConf, certConfFile, FilePerm)
	if err != nil {
		return fmt.Errorf("dump server_cert_conf: %v", err)
	}

	return nil
}

func (c *BfeTLSIngressConfig) Reload() error {
	if !isConfigEqual(ConfigNameTLSConf, c.tc) {
		updateLastConfig(ConfigNameTLSConf, c.tc)
		return reloadBfe(ConfigNameTLSConf)
	}
	return nil
}

func (c *BfeTLSIngressConfig) getCertFilePath(host string) string {
	return ConfigPath + CertKeyFilePath + host + ".cer"
}

func (c *BfeTLSIngressConfig) getKeyFilePath(host string) string {
	return ConfigPath + CertKeyFilePath + host + ".key"
}

func (c *BfeTLSIngressConfig) getcertConfFilePath() string {
	return ConfigPath + ServerCertData
}
