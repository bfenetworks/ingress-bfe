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

	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/host_rule_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
	"github.com/bfenetworks/ingress-bfe/internal/option"
)

const (
	DefaultProduct       = "default"
	ConfigNameServerData = "server_data_conf"
)

var (
	HostRuleData    = "server_data_conf/host_rule.data"
	RouteRuleData   = "server_data_conf/route_rule.data"
	ClusterConfData = "server_data_conf/cluster_conf.data"
)

type ServerDataConfig struct {
	hostTableVersion      string
	routeTableVersion     string
	bfeClusterConfVersion string

	routeRuleCache *RouteRuleCache

	hostTableConf  *host_rule_conf.HostTableConf
	routeTableFile *route_rule_conf.RouteTableFile
	bfeClusterConf *cluster_conf.BfeClusterConf
}

func NewServerDataConfig(version string) *ServerDataConfig {
	return &ServerDataConfig{
		routeRuleCache: newRouteRuleCache(version),
		hostTableConf:  newHostTableConf(version),
		routeTableFile: newRouteTableConfFile(version),
		bfeClusterConf: newBfeClusterConf(version),
	}
}

func newHostTableConf(version string) *host_rule_conf.HostTableConf {
	hostTagToHost := make(host_rule_conf.HostTagToHost)
	productToHostTag := make(host_rule_conf.ProductToHostTag)

	// all requests go to default product
	product := DefaultProduct
	var hostnameList host_rule_conf.HostnameList
	hostnameList = append(hostnameList, product)
	hostTagToHost[product] = &hostnameList

	list := host_rule_conf.HostTagList{product}
	productToHostTag[product] = &list

	return &host_rule_conf.HostTableConf{
		Version:        &version,
		DefaultProduct: &product,
		Hosts:          &hostTagToHost,
		HostTags:       &productToHostTag,
	}
}

// newRouteTableConfFile build route table for all ingress rules
func newRouteTableConfFile(version string) *route_rule_conf.RouteTableFile {
	basicRule := make(route_rule_conf.ProductBasicRouteRuleFile)
	productRule := make(route_rule_conf.ProductAdvancedRouteRuleFile)
	routeTable := &route_rule_conf.RouteTableFile{
		Version:     &version,
		BasicRule:   &basicRule,
		ProductRule: &productRule,
	}

	(*routeTable.BasicRule)[DefaultProduct] = make(route_rule_conf.BasicRouteRuleFiles, 0)
	(*routeTable.ProductRule)[DefaultProduct] = make(route_rule_conf.AdvancedRouteRuleFiles, 0)

	return routeTable
}

func newBfeClusterConf(version string) *cluster_conf.BfeClusterConf {
	clusterToConf := make(cluster_conf.ClusterToConf)
	clusterConf := cluster_conf.BfeClusterConf{
		Version: &version,
		Config:  &clusterToConf,
	}

	return &clusterConf
}

func (c *ServerDataConfig) UpdateIngress(ingress *netv1.Ingress) error {
	if len(ingress.Spec.Rules) == 0 {
		return nil
	}

	ingressName := util.NamespacedName(ingress.Namespace, ingress.Name)

	//delete existing ingress
	if c.routeRuleCache.ContainsIngress(ingressName) {
		c.routeRuleCache.DeleteByIngress(ingressName)
	}

	if err := c.updateCache(ingress); err != nil {
		// delete rules which have been inserted
		c.routeRuleCache.DeleteByIngress(ingressName)
		return err
	}

	// TODO: avoid calling c.updateRouteTable() and c.updateBfeClusterConf() frequently

	if err := c.updateRouteTable(); err != nil {
		c.routeRuleCache.DeleteByIngress(ingressName)
		return err
	}

	c.updateBfeClusterConf()

	return nil
}

func (c *ServerDataConfig) DeleteIngress(namespace, name string) {
	ingressName := util.NamespacedName(namespace, name)

	if !c.routeRuleCache.ContainsIngress(ingressName) {
		return
	}

	c.routeRuleCache.DeleteByIngress(ingressName)
	_ = c.updateRouteTable()
	c.updateBfeClusterConf()
}

func (c *ServerDataConfig) updateCache(ingress *netv1.Ingress) error {
	return c.routeRuleCache.UpdateByIngress(ingress)
}

func (c *ServerDataConfig) updateRouteTable() error {
	basicRules, advancedRules := c.routeRuleCache.getRouteRules()

	routeTableFile := newRouteTableConfFile(util.NewVersion())
	for _, rule := range basicRules {
		ruleFile := route_rule_conf.BasicRouteRuleFile{
			ClusterName: &rule.Cluster,
		}

		if len(rule.GetHost()) > 0 && rule.GetHost() != "*" {
			ruleFile.Hostname = []string{rule.GetHost()}
		}

		if len(rule.GetPath()) > 0 {
			ruleFile.Path = []string{rule.GetPath()}
		}

		(*routeTableFile.BasicRule)[DefaultProduct] = append(
			(*routeTableFile.BasicRule)[DefaultProduct], ruleFile)
	}

	for _, rule := range advancedRules {
		condition, err := rule.GetCond()
		if err != nil {
			return err
		}
		ruleFile := route_rule_conf.AdvancedRouteRuleFile{
			Cond:        &condition,
			ClusterName: &rule.Cluster,
		}
		(*routeTableFile.ProductRule)[DefaultProduct] = append((*routeTableFile.ProductRule)[DefaultProduct], ruleFile)
	}

	if len(option.Opts.Ingress.DefaultBackend) > 0 && (len(basicRules) > 0 || len(advancedRules) > 0) {
		condition := "default_t()"
		cluster := util.DefaultClusterName()
		ruleFile := route_rule_conf.AdvancedRouteRuleFile{
			Cond:        &condition,
			ClusterName: &cluster,
		}
		(*routeTableFile.ProductRule)[DefaultProduct] = append((*routeTableFile.ProductRule)[DefaultProduct], ruleFile)
	}

	// check routeTableFile
	if _, err := route_rule_conf.Convert(routeTableFile); err != nil {
		return fmt.Errorf("fail to check generated routeTableFile, err: %s", err)
	}

	c.routeTableFile = routeTableFile

	return nil
}

func (c *ServerDataConfig) updateBfeClusterConf() {
	basicRules, advancedRules := c.routeRuleCache.getRouteRules()

	clusterConf := newBfeClusterConf(util.NewVersion())

	for _, r := range basicRules {
		if r.Cluster == route_rule_conf.AdvancedMode {
			continue
		}
		(*clusterConf.Config)[r.Cluster] = cluster_conf.ClusterConf{
			CheckConf: newCheckConf(),
			GslbBasic: newGslbBasicConf(),
		}
	}

	for _, r := range advancedRules {
		(*clusterConf.Config)[r.Cluster] = cluster_conf.ClusterConf{
			CheckConf: newCheckConf(),
			GslbBasic: newGslbBasicConf(),
		}
	}
	if len(option.Opts.Ingress.DefaultBackend) > 0 && (len(basicRules) > 0 || len(advancedRules) > 0) {
		(*clusterConf.Config)[util.DefaultClusterName()] = cluster_conf.ClusterConf{
			CheckConf: newCheckConf(),
			GslbBasic: newGslbBasicConf(),
		}
	}

	c.bfeClusterConf = clusterConf
}

func newCheckConf() *cluster_conf.BackendCheck {
	schem := "tcp"
	return &cluster_conf.BackendCheck{
		Schem: &schem,
	}
}

func newGslbBasicConf() *cluster_conf.GslbBasicConf {
	defaultHashStrategy := cluster_conf.ClientIdOnly
	defaultHashHeader := "bfe-non-existence"
	defaultSessionSticky := false
	gslbConf := &cluster_conf.GslbBasicConf{
		HashConf: &cluster_conf.HashConf{
			HashStrategy:  &defaultHashStrategy,
			HashHeader:    &defaultHashHeader,
			SessionSticky: &defaultSessionSticky,
		},
	}
	return gslbConf
}

func (c *ServerDataConfig) Reload() error {
	reload := false
	if *c.hostTableConf.Version != c.hostTableVersion {
		err := util.DumpBfeConf(HostRuleData, c.hostTableConf)
		if err != nil {
			return fmt.Errorf("dump gslb.data error: %v", err)
		}
		reload = true
	}
	if *c.routeTableFile.Version != c.routeTableVersion {
		if err := c.updateRouteTable(); err != nil {
			if err != nil {
				return fmt.Errorf("dump cluster_table.data error: %v", err)
			}
		}
		err := util.DumpBfeConf(RouteRuleData, c.routeTableFile)
		if err != nil {
			return fmt.Errorf("dump cluster_table.data error: %v", err)
		}
		reload = true
	}

	if *c.bfeClusterConf.Version != c.bfeClusterConfVersion {
		c.updateBfeClusterConf()
		err := util.DumpBfeConf(ClusterConfData, c.bfeClusterConf)
		if err != nil {
			return fmt.Errorf("dump cluster_table.data error: %v", err)
		}
		reload = true
	}

	if reload {
		if err := util.ReloadBfe(ConfigNameServerData); err != nil {
			return err
		}
		c.hostTableVersion = *c.hostTableConf.Version
		c.routeTableVersion = *c.routeTableFile.Version
		c.bfeClusterConfVersion = *c.bfeClusterConf.Version
	}

	return nil
}
