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
package configs

import (
	"fmt"

	"github.com/bfenetworks/bfe/bfe_config/bfe_tls_conf/server_cert_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_tls_conf/tls_rule_conf"
	"github.com/bfenetworks/bfe/bfe_tls"
	"github.com/jwangsadinata/go-multimap/setmultimap"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

const (
	DefaultCNName = "example"
)

type certConf struct {
	cert []byte
	key  []byte
}

var (
	ConfigNameTLSConf = "tls_conf"

	ServerCertData  = "tls_conf/server_cert_conf.data"
	TLSRuleData     = "tls_conf/tls_rule_conf.data"
	CertKeyFilePath = "tls_conf/certs/"

	SecretCrt = "tls.crt"
	SecretKey = "tls.key"
)

type TLSConfig struct {
	serverCertVersion string
	tlsRuleVersion    string

	ingress2secret *setmultimap.MultiMap

	serverCertConf *server_cert_conf.BfeServerCertConf
	tlsRuleConf    *tls_rule_conf.BfeTlsRuleConf
	certs          map[string]certConf
}

func NewTLSConfig(version string) *TLSConfig {
	tlsConf := &TLSConfig{
		ingress2secret: setmultimap.New(),
		serverCertConf: newServerCertConf(version),
		tlsRuleConf:    newTlsRuleConf(version),
		certs:          make(map[string]certConf),
	}

	return tlsConf
}

func newServerCertConf(version string) *server_cert_conf.BfeServerCertConf {
	certConf := &server_cert_conf.BfeServerCertConf{
		Version: version,
		Config:  server_cert_conf.ServerCertConfMap{},
	}

	certConf.Config.Default = DefaultCNName
	certConf.Config.CertConf = make(map[string]server_cert_conf.ServerCertConf)
	certConf.Config.CertConf[DefaultCNName] =
		server_cert_conf.ServerCertConf{
			ServerCertFile:   getCertFilePath(DefaultCNName),
			ServerKeyFile:    getKeyFilePath(DefaultCNName),
			OcspResponseFile: "",
		}
	return certConf
}

func newTlsRuleConf(version string) *tls_rule_conf.BfeTlsRuleConf {
	ruleConf := &tls_rule_conf.BfeTlsRuleConf{
		Version:              version,
		Config:               make(tls_rule_conf.TlsRuleMap),
		DefaultNextProtos:    nil,
		DefaultChacha20:      false,
		DefaultDynamicRecord: false,
	}
	return ruleConf
}

func (c *TLSConfig) setVersion() {
	version := util.NewVersion()

	c.serverCertConf.Version = version
	c.tlsRuleConf.Version = version
}

func (c *TLSConfig) UpdateIngress(ingress *netv1.Ingress, secrets []*corev1.Secret) error {
	ingressName := util.NamespacedName(ingress.Namespace, ingress.Name)
	for _, secret := range secrets {
		secretName := util.NamespacedName(secret.Namespace, secret.Name)
		c.ingress2secret.Put(ingressName, secretName)
	}

	for _, secret := range secrets {
		if err := c.UpdateSecret(secret); err != nil {
			return err
		}
	}

	return nil
}

func (c *TLSConfig) DeleteIngress(namespace, name string) {
	ingressName := util.NamespacedName(namespace, name)
	if !c.ingress2secret.ContainsKey(ingressName) {
		return
	}

	c.ingress2secret.RemoveAll(ingressName)

	// check all certs to find which one should be deleted
	updated := false
	secrets := c.ingress2secret.Values()
	for name := range c.certs {
		found := false
		for _, secret := range secrets {
			secretName := secret.(string)
			if secretName == name {
				// not used anymore, delete it
				found = true
			}
		}
		if !found {
			// not used anymore, delete it
			c.deleteCert(name)
			updated = true
		}
	}
	if updated {
		c.setVersion()
	}
}

func (c *TLSConfig) deleteCert(name string) {
	if cert, ok := c.serverCertConf.Config.CertConf[name]; ok {
		util.DeleteFile(cert.ServerKeyFile)
		util.DeleteFile(cert.ServerCertFile)
	}
	delete(c.serverCertConf.Config.CertConf, name)
	delete(c.certs, name)
}

func (c *TLSConfig) UpdateSecret(secret *corev1.Secret) error {
	name := util.NamespacedName(secret.Namespace, secret.Name)
	if !c.ingress2secret.ContainsValue(name) {
		return nil
	}

	_, err := bfe_tls.X509KeyPair(secret.Data[SecretCrt], secret.Data[SecretKey])
	if err != nil {
		return err
	}

	serverCertConf := server_cert_conf.ServerCertConf{
		ServerCertFile:   getCertFilePath(name),
		ServerKeyFile:    getKeyFilePath(name),
		OcspResponseFile: "",
	}

	c.serverCertConf.Config.CertConf[name] = serverCertConf
	c.certs[name] = certConf{
		cert: secret.Data[SecretCrt],
		key:  secret.Data[SecretKey],
	}

	c.setVersion()

	return nil
}

func (c *TLSConfig) DeleteSecret(namespace, name string) {
	target := util.NamespacedName(namespace, name)

	if !c.ingress2secret.ContainsValue(target) {
		return
	}

	c.deleteCert(target)

	c.setVersion()
}

func (c *TLSConfig) Reload() error {
	reload := false
	if c.serverCertConf.Version != c.serverCertVersion {
		err := util.DumpBfeConf(ServerCertData, c.serverCertConf)
		if err != nil {
			return fmt.Errorf("dump server_cert_conf: %v", err)
		}

		for name, cert := range c.serverCertConf.Config.CertConf {
			if name == DefaultCNName {
				continue
			}
			if err = util.DumpFile(cert.ServerCertFile, c.certs[name].cert); err != nil {
				return err
			}
			if err = util.DumpFile(cert.ServerKeyFile, c.certs[name].key); err != nil {
				return err
			}
		}
		reload = true
	}

	if c.tlsRuleConf.Version != c.tlsRuleVersion {
		err := util.DumpBfeConf(TLSRuleData, c.tlsRuleConf)
		if err != nil {
			return fmt.Errorf("dump tls_rule_conf: %v", err)
		}
		reload = true
	}

	if reload {
		if err := util.ReloadBfe(ConfigNameTLSConf); err != nil {
			return err
		}
		c.serverCertVersion = c.serverCertConf.Version
		c.tlsRuleVersion = c.tlsRuleConf.Version
	}

	return nil
}

func getCertFilePath(name string) string {
	return CertKeyFilePath + name + ".crt"
}

func getKeyFilePath(name string) string {
	return CertKeyFilePath + name + ".key"
}
