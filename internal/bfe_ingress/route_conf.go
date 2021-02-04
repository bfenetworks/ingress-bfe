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
)

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/host_rule_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/bfenetworks/bfe/bfe_util"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

const (
	DefaultProduct      = "default"
	ConfigNameRouteConf = "server_data_conf"

	UnknownConditionType = -1

	ConditionTypeContainHostExactPath = iota
	ConditionTypeContainHostPrefixPath
	ConditionTypeContainOnlyHost
	ConditionTypeContainOnlyExactPath
	ConditionTypeContainOnlyPrefixPath
	ConditionTypeContainNoHostPath
)

type BfeRouteConf struct {
	hostTableConf  *host_rule_conf.HostTableConf
	routeTableFile *route_rule_conf.RouteTableFile
	bfeClusterConf *cluster_conf.BfeClusterConf
}

type ingressRawRuleInfo struct {
	Host        string
	Path        string
	Annotations []BfeAnnotation
}

type ingressRouteRuleFile struct {
	RouteRuleFile route_rule_conf.RouteRuleFile
	RawRuleInfo   ingressRawRuleInfo
	ConditionType int
}

var (
	HostRuleData    = "server_data_conf/host_rule.data"
	RouteRuleData   = "server_data_conf/route_rule.data"
	ClusterConfData = "server_data_conf/cluster_conf.data"
)

type ingressRecordRule struct {
	rule    *ingressRouteRuleFile
	ingress *networkingv1beta1.Ingress
}

// record condition str -> ingress record rule
type Rule map[string]*ingressRecordRule

// record host-> route config of this host
type HostRule map[string]Rule

type BfeRouteIngressConfig struct {
	client  *kubernetes_client.KubernetesClient
	version string

	routeConf BfeRouteConf

	rules HostRule
}

func NewBfeRouteIngressConfig(client *kubernetes_client.KubernetesClient, version string) *BfeRouteIngressConfig {
	c := &BfeRouteIngressConfig{}
	c.client = client
	c.version = version
	c.rules = make(HostRule)
	return c
}

func (c *BfeRouteIngressConfig) Submit(ingress *networkingv1beta1.Ingress) error {
	var balance LoadBalance
	var err error
	for key, value := range ingress.Annotations {
		if key == LoadBalanceAnnotation {
			balance, err = BuildLoadBalanceAnnotation(key, value)
			if err != nil {
				return err
			}
			break
		}
	}

	annotationConds := make([]BfeAnnotation, 0)
	annotations := ingress.Annotations
	for key, value := range annotations {
		bfeAnnotation, err := BuildBfeAnnotation(key, value)
		if err != nil {
			// log.Logger.Debug("buildRouteTableConfFile() error. key[%s], err[%v]", key, err.Error())
			continue
		}
		annotationConds = append(annotationConds, bfeAnnotation)
	}
	SortAnnotations(annotationConds)
	var cacheHostRule = make(HostRule)
	for _, rule := range ingress.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			var clusterName string
			if !balance.ContainService(p.Backend.ServiceName) {
				clusterName = GetSingleClusterName(ingress.Namespace, p.Backend.ServiceName)
			} else {
				clusterName = GetMultiClusterName(ingress.Namespace, ingress.Name, p.Backend.ServiceName)
			}

			cond, conditionType := buildCondition(rule.Host, p.Path, p.PathType, annotationConds)
			routeRuleFileVal := ingressRouteRuleFile{
				RouteRuleFile: route_rule_conf.RouteRuleFile{
					Cond:        &cond,
					ClusterName: &clusterName,
				},
				RawRuleInfo: ingressRawRuleInfo{
					Host:        rule.Host,
					Path:        p.Path,
					Annotations: annotationConds,
				},
				ConditionType: conditionType,
			}
			product := DefaultProduct
			if rule.Host != "" {
				product = rule.Host
			}
			ruleRecord := &ingressRecordRule{
				rule:    &routeRuleFileVal,
				ingress: ingress,
			}
			if _, ok := c.rules[product]; !ok {
				c.rules[product] = make(Rule)
			}
			if _, ok := c.rules[product][cond]; ok {
				conflictIngress := c.rules[product][cond].ingress
				msg := fmt.Sprintf("route cond conflict, ingress[%s/%s] ingored cause other ingress[%s/%s]", ingress.Namespace, ingress.Name,
					conflictIngress.Namespace, conflictIngress.Name)
				return fmt.Errorf(msg)
			}
			if _, ok := cacheHostRule[product]; !ok {
				cacheHostRule[product] = make(Rule)
			}
			if _, ok := cacheHostRule[product][cond]; ok {
				conflictIngress := cacheHostRule[product][cond].ingress
				msg := fmt.Sprintf("route cond conflict, ingress[%s/%s] ingored cause other ingress[%s/%s]", ingress.Namespace, ingress.Name,
					conflictIngress.Namespace, conflictIngress.Name)
				return fmt.Errorf(msg)
			}
			cacheHostRule[product][cond] = ruleRecord
		}
	}
	for product, productRule := range cacheHostRule {
		if _, ok := c.rules[product]; !ok {
			c.rules[product] = make(Rule)
		}
		for cond, rule := range productRule {
			c.rules[product][cond] = rule
		}
	}

	return nil
}

func (c *BfeRouteIngressConfig) Rollback(ingress *networkingv1beta1.Ingress) error {

	annotationConds := make([]BfeAnnotation, 0)
	annotations := ingress.Annotations
	if len(annotations) != 0 {
		for key, value := range annotations {
			bfeAnnotation, err := BuildBfeAnnotation(key, value)
			if err != nil {
				// log.Logger.Debug("buildRouteTableConfFile() error. key[%s], err[%v]", key, err.Error())
				continue
			}
			annotationConds = append(annotationConds, bfeAnnotation)
		}
	}
	SortAnnotations(annotationConds)

	for _, rule := range ingress.Spec.Rules {
		for _, p := range rule.HTTP.Paths {

			cond, _ := buildCondition(rule.Host, p.Path, p.PathType, annotationConds)

			product := DefaultProduct
			if rule.Host != "" {
				product = rule.Host
			}

			if _, ok := c.rules[product]; !ok {
				return fmt.Errorf("Rollback unknown product")
			}
			//Note: cause cond cannot repeated, so we donot need to judge refCount in routeRule
			if _, ok := c.rules[product][cond]; ok {
				delete(c.rules[product], cond)
			}
		}
	}

	return nil
}

func (c *BfeRouteIngressConfig) Build() error {
	clusterConf, err := c.buildBfeClusterConf()
	if err != nil {
		return err
	}
	hostConf := c.buildHostTableConf()
	route, err := c.buildRouteTableConfFile()
	if err != nil {
		return err
	}
	c.routeConf = BfeRouteConf{
		hostTableConf:  &hostConf,
		routeTableFile: &route,
		bfeClusterConf: &clusterConf,
	}
	return nil
}

func (c *BfeRouteIngressConfig) buildBfeClusterConf() (cluster_conf.BfeClusterConf, error) {

	clusterToConf := make(cluster_conf.ClusterToConf)

	clusterConfs := cluster_conf.BfeClusterConf{
		Version: &c.version,
		Config:  &clusterToConf,
	}

	for _, rules := range c.rules {
		for _, rule := range rules {
			clusterName := rule.rule.RouteRuleFile.ClusterName
			gslbConf := InitClusterGslb()
			clusterToConf[*clusterName] = cluster_conf.ClusterConf{GslbBasic: gslbConf}
		}
	}

	err := cluster_conf.ClusterToConfCheck(clusterToConf)
	if err != nil {
		return clusterConfs, err
	}

	return clusterConfs, nil
}

func (c *BfeRouteIngressConfig) buildHostTableConf() host_rule_conf.HostTableConf {
	hostTagToHost := make(host_rule_conf.HostTagToHost)
	productToHostTag := make(host_rule_conf.ProductToHostTag)
	product := DefaultProduct
	defaultProduct := DefaultProduct
	for host := range c.rules {
		product = host
		var hostnameList host_rule_conf.HostnameList
		hostnameList = append(hostnameList, host)
		hostTagToHost[product] = &hostnameList

		list := host_rule_conf.HostTagList{product}
		productToHostTag[product] = &list
	}

	defaultHostList := host_rule_conf.HostnameList{defaultProduct}
	defaultProductList := host_rule_conf.HostTagList{defaultProduct}
	hostTagToHost[defaultProduct] = &defaultHostList
	productToHostTag[defaultProduct] = &defaultProductList

	return host_rule_conf.HostTableConf{
		Version:        &c.version,
		DefaultProduct: &defaultProduct,
		Hosts:          &hostTagToHost,
		HostTags:       &productToHostTag,
	}
}

func (c *BfeRouteIngressConfig) buildRouteTableConfFile() (route_rule_conf.RouteTableFile, error) {
	var routeTable route_rule_conf.RouteTableFile
	var p route_rule_conf.ProductRouteRuleFile = make(route_rule_conf.ProductRouteRuleFile)
	routeTable.ProductRule = &p
	for host, rules := range c.rules {
		var routeRuleFiles []ingressRouteRuleFile
		for _, rule := range rules {
			routeRuleFiles = append(routeRuleFiles, *rule.rule)
		}
		sortRules(routeRuleFiles)
		var rules route_rule_conf.RouteRuleFiles
		for _, routeRuleFile := range routeRuleFiles {
			rules = append(rules, route_rule_conf.RouteRuleFile{
				Cond:        routeRuleFile.RouteRuleFile.Cond,
				ClusterName: routeRuleFile.RouteRuleFile.ClusterName,
			})
		}
		(*routeTable.ProductRule)[host] = rules
	}
	routeTable.Version = &c.version
	return routeTable, nil

}

func (c *BfeRouteIngressConfig) Dump() error {
	err := bfe_util.DumpJson(c.routeConf.hostTableConf, ConfigPath+HostRuleData, FilePerm)
	if err != nil {
		return fmt.Errorf("dump host_rule.data error: %v", err)
	}

	err = bfe_util.DumpJson(c.routeConf.routeTableFile, ConfigPath+RouteRuleData, FilePerm)
	if err != nil {
		return fmt.Errorf("dump route_rule.data error: %v", err)
	}

	err = bfe_util.DumpJson(c.routeConf.bfeClusterConf, ConfigPath+ClusterConfData, FilePerm)
	if err != nil {
		return fmt.Errorf("dump cluster_conf.data error: %v", err)
	}

	return nil
}

func (c *BfeRouteIngressConfig) Reload() error {
	if !isConfigEqual(ConfigNameRouteConf, c.routeConf) {
		updateLastConfig(ConfigNameRouteConf, c.routeConf)
		return reloadBfe(ConfigNameRouteConf)
	}
	return nil
}

func buildCondition(host string, path string, pathType *networkingv1beta1.PathType, exConds []BfeAnnotation) (string, int) {
	condType := UnknownConditionType
	var cond string
	if len(host) != 0 && len(path) != 0 {
		if pathType != nil && *pathType == networkingv1beta1.PathTypeExact {
			cond = fmt.Sprintf("req_host_in(\"%s\") && req_path_in(\"%s\", false)", host, path)
			condType = ConditionTypeContainHostExactPath
		} else {
			cond = fmt.Sprintf("req_host_in(\"%s\") && req_path_element_prefix_in(\"%s\", false)", host, path)
			condType = ConditionTypeContainHostPrefixPath
		}
	}
	if condType == UnknownConditionType && len(host) != 0 {
		cond = fmt.Sprintf("req_host_in(\"%s\")", host)
		condType = ConditionTypeContainOnlyHost
	}
	if condType == UnknownConditionType && len(path) != 0 {
		if pathType != nil && *pathType == networkingv1beta1.PathTypeExact {
			cond = fmt.Sprintf("req_path_in(\"%s\", false)", path)
			condType = ConditionTypeContainOnlyExactPath
		} else {
			cond = fmt.Sprintf("req_path_element_prefix_in(\"%s\", false)", path)
			condType = ConditionTypeContainOnlyPrefixPath
		}
	}

	if condType == UnknownConditionType {
		cond = fmt.Sprintf("default_t()")
		return cond, ConditionTypeContainNoHostPath
	}

	for _, exCond := range exConds {
		cond = fmt.Sprintf("%s && %s", cond, exCond.Build())
	}
	return cond, condType
}

func sortRules(routeRuleFiles []ingressRouteRuleFile) {
	sort.Slice(routeRuleFiles, func(i, j int) bool {
		// Sort by ConditionType.
		// As ContainHostPath > ContainOnlyHost > ContainOnlyPath > ContainNoHostPath
		if routeRuleFiles[i].ConditionType != routeRuleFiles[j].ConditionType {
			return routeRuleFiles[i].ConditionType < routeRuleFiles[j].ConditionType
		}

		// Sort by Path length if ConditionType is same, more exact path with higher weight;
		// as path(/api/v1/route) with higher weight than path(/api/v1)
		if len(routeRuleFiles[i].RawRuleInfo.Path) != len(routeRuleFiles[j].RawRuleInfo.Path) {
			return len(routeRuleFiles[i].RawRuleInfo.Path) > len(routeRuleFiles[j].RawRuleInfo.Path)
		}

		// Sort by length of annotations if the path is same;
		// as condition with header and cookie with higher weight than header;
		if len(routeRuleFiles[i].RawRuleInfo.Annotations) != len(routeRuleFiles[j].RawRuleInfo.Annotations) {
			return len(routeRuleFiles[i].RawRuleInfo.Annotations) > len(routeRuleFiles[j].RawRuleInfo.Annotations)
		}

		// Sort by annotations's Priority as length of annotations is same;
		// for example, cookie is greater than header;
		for index := 0; index < len(routeRuleFiles[i].RawRuleInfo.Annotations); index++ {
			if routeRuleFiles[i].RawRuleInfo.Annotations[index].Priority() != routeRuleFiles[j].RawRuleInfo.Annotations[index].Priority() {
				return routeRuleFiles[i].RawRuleInfo.Annotations[index].Priority() < routeRuleFiles[j].RawRuleInfo.Annotations[index].Priority()
			}
		}

		// Sort by length if all above is same.
		if len(*routeRuleFiles[i].RouteRuleFile.Cond) != len(*routeRuleFiles[j].RouteRuleFile.Cond) {
			return len(*routeRuleFiles[i].RouteRuleFile.Cond) > len(*routeRuleFiles[j].RouteRuleFile.Cond)
		}

		// Sort by condition if all above is same.
		if *routeRuleFiles[i].RouteRuleFile.Cond != *routeRuleFiles[j].RouteRuleFile.Cond {
			return *routeRuleFiles[i].RouteRuleFile.Cond > *routeRuleFiles[j].RouteRuleFile.Cond
		}

		// Sort by cluster name if condition is same.
		return *routeRuleFiles[i].RouteRuleFile.ClusterName > *routeRuleFiles[j].RouteRuleFile.ClusterName
	})
}
