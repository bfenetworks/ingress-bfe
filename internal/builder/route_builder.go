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
	"path"
	"sort"
	"strings"
)

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/host_rule_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/utils/pointer"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

const (
	DefaultProduct      = "default"
	ConfigNameRouteConf = "server_data_conf"

	UnknownConditionType = -1

	ConditionTypeContainExactHostExactPath = iota
	ConditionTypeContainExactHostPrefixPath
	ConditionTypeContainOnlyExactHost
	ConditionTypeContainWildcardHostExactPath
	ConditionTypeContainWildcardHostPrefixPath
	ConditionTypeContainOnlyWildcardHost
	ConditionTypeContainOnlyExactPath
	ConditionTypeContainOnlyPrefixPath
	ConditionTypeContainNoHostPath
)

const (
	ClusterNameAdvancedMode = "ADVANCED_MODE"
)

const (
	HostTypeExact = iota
	HostTypeWildcard
	HostTypeNoRestriction
)

type BfeRouteConf struct {
	hostTableConf  *host_rule_conf.HostTableConf
	routeTableFile *route_rule_conf.RouteTableFile
	bfeClusterConf *cluster_conf.BfeClusterConf
}

type ingressRawRuleInfo struct {
	Host        string
	Path        string
	PathType    *networking.PathType
	Annotations []BfeAnnotation
}

type ingressRouteRuleFile struct {
	RouteRuleFile route_rule_conf.AdvancedRouteRuleFile
	RawRuleInfo   ingressRawRuleInfo
	ConditionType int
}

const (
	ServerDataConfDir = "server_data_conf/"

	HostRuleData    = ServerDataConfDir + "host_rule.data"
	RouteRuleData   = ServerDataConfDir + "route_rule.data"
	ClusterConfData = ServerDataConfDir + "cluster_conf.data"
)

type ingressRecordRule struct {
	rule    *ingressRouteRuleFile
	ingress *networking.Ingress
}

// record condition str -> ingress record rule
type Rule map[string]*ingressRecordRule

// record host-> route config of this host
type HostRule map[string]Rule

// Get rule by host and condition
func (r HostRule) Get(host, condition string) (*ingressRecordRule, bool) {
	conditionRule, ok := r[host]
	if !ok {
		return nil, false
	}

	if rule, ok := conditionRule[condition]; !ok {
		return nil, false
	} else {
		return rule, true
	}
}

// DisposableSet sets rule for host and condition, set only success once for same host and condition
// previous set value will returned when failed
func (r HostRule) DisposableSet(host, condition string, value *ingressRecordRule) (*ingressRecordRule, bool) {
	conditionRule, ok := r[host]
	if !ok {
		conditionRule = make(Rule)
		r[host] = conditionRule
	}

	if rule, ok := conditionRule[condition]; !ok {
		conditionRule[condition] = value
		return nil, true
	} else {
		return rule, false
	}
}

type BfeRouteConfigBuilder struct {
	client *kubernetes_client.KubernetesClient

	dumper   *Dumper
	reloader *Reloader
	version  string

	routeConf BfeRouteConf

	rules HostRule
}

// advancedRuleCoverage to record host & path of current advanced rules
type advancedRuleCoverage struct {
	HostPath map[string]map[string]bool
}

func NewBfeRouteConfigBuilder(client *kubernetes_client.KubernetesClient, version string, dumper *Dumper, r *Reloader) *BfeRouteConfigBuilder {
	c := &BfeRouteConfigBuilder{}
	c.client = client
	c.version = version
	c.dumper = dumper
	c.reloader = r
	c.rules = make(HostRule)
	return c
}

func (c *BfeRouteConfigBuilder) Submit(ingress *networking.Ingress) error {
	// build balance from annotation
	var balance LoadBalance
	var err error
	for key, value := range ingress.Annotations {
		if key == LoadBalanceWeightAnnotation {
			balance, err = BuildLoadBalanceAnnotation(key, value)
			if err != nil {
				return err
			}
			break
		}
	}

	// generate rules in cache
	var cacheHostRule = make(HostRule)
	annotationConds := BuildBfeAnnotations(ingress.Annotations)
	for _, rule := range ingress.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			product := rule.Host
			if product == "" {
				product = DefaultProduct
			}
			cond, conditionType := buildCondition(rule.Host, p.Path, p.PathType, annotationConds)

			// check conflict with previous rules in previous ingress
			if conflictRule, ok := c.rules.Get(product, cond); ok {
				conflictIngress := conflictRule.ingress
				return fmt.Errorf("route cond conflict, ingress[%s/%s] ingored cause other ingress[%s/%s]",
					ingress.Namespace, ingress.Name, conflictIngress.Namespace, conflictIngress.Name)
			}

			// generate rule and add to cache
			ruleRecord := recordRule(ingress, rule, cond, conditionType, balance, annotationConds, p)
			if conflictRule, ok := cacheHostRule.DisposableSet(product, cond, ruleRecord); !ok {
				conflictIngress := conflictRule.ingress
				return fmt.Errorf("route cond conflict, ingress[%s/%s] ingored cause other ingress[%s/%s]",
					ingress.Namespace, ingress.Name, conflictIngress.Namespace, conflictIngress.Name)
			}
		}
	}

	// save rules from cache to ConfigBuilder
	for product, productRule := range cacheHostRule {
		for cond, rule := range productRule {
			c.rules.DisposableSet(product, cond, rule)
		}
	}

	return nil
}

func recordRule(ingress *networking.Ingress, rule networking.IngressRule, cond string, conditionType int,
	balance LoadBalance, annotations []BfeAnnotation, p networking.HTTPIngressPath) *ingressRecordRule {
	clusterName := ClusterName(ingress, balance, p)
	ruleFile := ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
			Cond:        &cond,
			ClusterName: &clusterName,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host:        rule.Host,
			Path:        p.Path,
			PathType:    p.PathType,
			Annotations: annotations,
		},
		ConditionType: conditionType,
	}
	ingressRecordRule := &ingressRecordRule{
		rule:    &ruleFile,
		ingress: ingress,
	}
	return ingressRecordRule
}

func (c *BfeRouteConfigBuilder) Rollback(ingress *networking.Ingress) error {
	annotationConds := BuildBfeAnnotations(ingress.Annotations)

	for _, rule := range ingress.Spec.Rules {
		for _, p := range rule.HTTP.Paths {

			cond, _ := buildCondition(rule.Host, p.Path, p.PathType, annotationConds)

			product := DefaultProduct
			if rule.Host != "" {
				product = rule.Host
			}

			if _, ok := c.rules[product]; !ok {
				return fmt.Errorf("rollback unknown product")
			}
			// Note: cause cond cannot repeated, so we do not need to judge refCount in routeRule
			if _, ok := c.rules[product][cond]; ok {
				delete(c.rules[product], cond)
			}
		}
	}

	return nil
}

func (c *BfeRouteConfigBuilder) Build() error {
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

func (c *BfeRouteConfigBuilder) buildBfeClusterConf() (cluster_conf.BfeClusterConf, error) {
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

func (c *BfeRouteConfigBuilder) buildHostTableConf() host_rule_conf.HostTableConf {
	hostTagToHost := make(host_rule_conf.HostTagToHost)
	productToHostTag := make(host_rule_conf.ProductToHostTag)

	// build for default product
	defaultProduct := DefaultProduct
	defaultHostList := host_rule_conf.HostnameList{defaultProduct}
	hostTagToHost[defaultProduct] = &defaultHostList
	defaultProductList := host_rule_conf.HostTagList{defaultProduct}
	productToHostTag[defaultProduct] = &defaultProductList

	// build for custom product
	for host := range c.rules {
		product := host
		hostnameList := host_rule_conf.HostnameList{host}
		hostTagToHost[product] = &hostnameList
		list := host_rule_conf.HostTagList{product}
		productToHostTag[product] = &list
	}

	return host_rule_conf.HostTableConf{
		Version:        &c.version,
		DefaultProduct: &defaultProduct,
		Hosts:          &hostTagToHost,
		HostTags:       &productToHostTag,
	}
}

// buildRouteTableConfFile builds route table for all product from ingress rules
func (c *BfeRouteConfigBuilder) buildRouteTableConfFile() (route_rule_conf.RouteTableFile, error) {
	var routeTable route_rule_conf.RouteTableFile
	productBasicRouteRule := make(route_rule_conf.ProductBasicRouteRuleFile)
	routeTable.BasicRule = &productBasicRouteRule
	productAdvancedRouteRule := make(route_rule_conf.ProductAdvancedRouteRuleFile)
	routeTable.ProductRule = &productAdvancedRouteRule

	cov := newAdvancedRuleCoverage()
	for host, rules := range c.rules {
		// collect rules
		var routeRuleFiles []ingressRouteRuleFile
		for _, rule := range rules {
			routeRuleFiles = append(routeRuleFiles, *rule.rule)
		}

		// sort rules
		sortRules(routeRuleFiles)

		// build rules and save to routeTable
		for _, routeRuleFile := range routeRuleFiles {
			buildRouteRule(host, routeRuleFile, cov, routeTable)
		}
	}
	routeTable.Version = &c.version
	return routeTable, nil
}

/*
buildRouteRule builds route rules, it's stateful.

Route rule is built according to current host, rule file and coverage,
and append built rule to result RouteTableFile
*/
func buildRouteRule(host string, ruleFile ingressRouteRuleFile, cov *advancedRuleCoverage,
	result route_rule_conf.RouteTableFile) {
	// basic route rule can't satisfied ingress with advanced BFE annotation
	if len(ruleFile.RawRuleInfo.Annotations) <= 0 {
		buildBasicRouteRule(host, ruleFile, cov, result)
	} else {
		// update advanced rule coverage
		cov.Cover(ruleFile.RawRuleInfo)

		buildProductRouteRule(host, ruleFile, result)
	}
}

/*
buildBasicRouteRule builds basic route rules, it's stateful.

Basic route rule is built according to current host, rule file and coverage,
and append built rule to result RouteTableFile.

If current basic route rule is covered by previous product route rule,
it will convert to advanced mode, and corresponding new product route rule is appended.
*/
func buildBasicRouteRule(host string, ruleFile ingressRouteRuleFile, cov *advancedRuleCoverage,
	result route_rule_conf.RouteTableFile) {
	basicRule := newBasicRouteRuleFile(ruleFile)

	if cov.IsCovered(ruleFile.RawRuleInfo) {
		// convert to advanced mode if covered by any advanced rule
		basicRule.ClusterName = pointer.StringPtr(ClusterNameAdvancedMode)
		buildProductRouteRule(host, ruleFile, result)
	}

	(*result.BasicRule)[host] = append((*result.BasicRule)[host], basicRule)
}

/*
buildProductRouteRule builds advanced route rules, it's stateful.

Product route rule is built according to current host and rule file,
and append built rule to result RouteTableFile
*/
func buildProductRouteRule(host string, ruleFile ingressRouteRuleFile, result route_rule_conf.RouteTableFile) {
	advancedRule := route_rule_conf.AdvancedRouteRuleFile{
		Cond:        ruleFile.RouteRuleFile.Cond,
		ClusterName: ruleFile.RouteRuleFile.ClusterName,
	}
	(*result.ProductRule)[host] = append((*result.ProductRule)[host], advancedRule)
}

func (c *BfeRouteConfigBuilder) Dump() error {
	err := c.dumper.DumpJson(c.routeConf.hostTableConf, HostRuleData)
	if err != nil {
		return fmt.Errorf("dump %s error: %v", HostRuleData, err)
	}

	err = c.dumper.DumpJson(c.routeConf.routeTableFile, RouteRuleData)
	if err != nil {
		return fmt.Errorf("dump %s error: %v", RouteRuleData, err)
	}

	err = c.dumper.DumpJson(c.routeConf.bfeClusterConf, ClusterConfData)
	if err != nil {
		return fmt.Errorf("dump %s error: %v", ClusterConfData, err)
	}

	return nil
}

func (c *BfeRouteConfigBuilder) Reload() error {
	return c.reloader.DoReload(c.routeConf, ConfigNameRouteConf)
}

func buildCondition(host string, path string, pathType *networking.PathType, exConds []BfeAnnotation) (string, int) {
	condType := UnknownConditionType
	bfePathType := BfePathType(pathType)

	stmts := make([]string, 0)

	hostStmt, hostType := hostStatement(host)
	pathStmt := pathStatement(path, bfePathType)
	stmts = append(stmts, hostStmt, pathStmt)

	// set condition type
	switch hostType {
	case HostTypeNoRestriction:
		if len(path) == 0 {
			return expression(stmts), ConditionTypeContainNoHostPath
		}

		switch bfePathType {
		case networking.PathTypeExact:
			condType = ConditionTypeContainOnlyExactPath
		default:
			condType = ConditionTypeContainOnlyPrefixPath
		}

	case HostTypeWildcard:
		if len(path) == 0 {
			condType = ConditionTypeContainOnlyWildcardHost
			break
		}

		switch bfePathType {
		case networking.PathTypeExact:
			condType = ConditionTypeContainWildcardHostExactPath
		default:
			condType = ConditionTypeContainWildcardHostPrefixPath
		}

	case HostTypeExact:
		if len(path) == 0 {
			condType = ConditionTypeContainOnlyExactHost
			break
		}

		switch bfePathType {
		case networking.PathTypeExact:
			condType = ConditionTypeContainExactHostExactPath
		default:
			condType = ConditionTypeContainExactHostPrefixPath
		}
	}

	for _, exCond := range exConds {
		stmts = append(stmts, exCond.Build())
	}
	return expression(stmts), condType
}

func sortRules(routeRuleFiles []ingressRouteRuleFile) {
	sort.Slice(routeRuleFiles, func(i, j int) bool {
		// Sort by ConditionType.
		// As ContainHostPath > ContainOnlyHost > ContainOnlyPath > ContainNoHostPath
		if routeRuleFiles[i].ConditionType != routeRuleFiles[j].ConditionType {
			return routeRuleFiles[i].ConditionType < routeRuleFiles[j].ConditionType
		}

		// Sort by Host length if ConditionType is same, more exact host with higher weight;
		// as host(www.baidu.com) with higher weight than host(baidu.com)
		if len(routeRuleFiles[i].RawRuleInfo.Host) != len(routeRuleFiles[j].RawRuleInfo.Host) {
			return len(routeRuleFiles[i].RawRuleInfo.Host) > len(routeRuleFiles[j].RawRuleInfo.Host)
		}

		// Sort by Path if Host length is same, more exact path with higher weight;
		// as path(/api/v1/route) with higher weight than path(/api/v1)
		if len(routeRuleFiles[i].RawRuleInfo.Path) != len(routeRuleFiles[j].RawRuleInfo.Path) {
			return len(routeRuleFiles[i].RawRuleInfo.Path) > len(routeRuleFiles[j].RawRuleInfo.Path)
		}

		// Sort by quantity of annotations if the path is same;
		// as condition with header and cookie with higher weight than header;
		if len(routeRuleFiles[i].RawRuleInfo.Annotations) != len(routeRuleFiles[j].RawRuleInfo.Annotations) {
			return len(routeRuleFiles[i].RawRuleInfo.Annotations) > len(routeRuleFiles[j].RawRuleInfo.Annotations)
		}

		// Sort by each annotation's Priority as quantity of annotations is same;
		// for example, cookie is greater than header;
		for index := 0; index < len(routeRuleFiles[i].RawRuleInfo.Annotations); index++ {
			if routeRuleFiles[i].RawRuleInfo.Annotations[index].Priority() != routeRuleFiles[j].RawRuleInfo.Annotations[index].Priority() {
				return routeRuleFiles[i].RawRuleInfo.Annotations[index].Priority() < routeRuleFiles[j].RawRuleInfo.Annotations[index].Priority()
			}
		}

		// Sort by length of condition if all above is same.
		if len(*routeRuleFiles[i].RouteRuleFile.Cond) != len(*routeRuleFiles[j].RouteRuleFile.Cond) {
			return len(*routeRuleFiles[i].RouteRuleFile.Cond) > len(*routeRuleFiles[j].RouteRuleFile.Cond)
		}

		// Sort by content of condition if all above is same.
		if *routeRuleFiles[i].RouteRuleFile.Cond != *routeRuleFiles[j].RouteRuleFile.Cond {
			return *routeRuleFiles[i].RouteRuleFile.Cond > *routeRuleFiles[j].RouteRuleFile.Cond
		}

		// Sort by cluster name if condition is same.
		return *routeRuleFiles[i].RouteRuleFile.ClusterName > *routeRuleFiles[j].RouteRuleFile.ClusterName
	})
}

func InitClusterGslb() *cluster_conf.GslbBasicConf {
	gslbConf := &cluster_conf.GslbBasicConf{}

	defaultRetryMax := 2
	gslbConf.RetryMax = &defaultRetryMax
	defaultCrossRetry := 0
	gslbConf.CrossRetry = &defaultCrossRetry

	defaultHashStrategy := cluster_conf.ClientIpOnly
	defaultSessionSticky := false
	gslbConf.HashConf = &cluster_conf.HashConf{
		HashStrategy:  &defaultHashStrategy,
		SessionSticky: &defaultSessionSticky,
	}

	defaultBalMode := cluster_conf.BalanceModeWrr
	gslbConf.BalanceMode = &defaultBalMode
	return gslbConf
}

func newBasicRouteRuleFile(rule ingressRouteRuleFile) route_rule_conf.BasicRouteRuleFile {
	return route_rule_conf.BasicRouteRuleFile{
		Hostname:    []string{rule.RawRuleInfo.Host},
		Path:        []string{rule.RawRuleInfo.GetPathPattern()},
		ClusterName: rule.RouteRuleFile.ClusterName,
	}
}

// expression builds final expression from statements with AND logic
// 		empty statement is allowed, and it will be ignored;
// 		if no valuable statement is provided, return default_t()
func expression(stmts []string) string {
	expressions := make([]string, 0)
	for _, stmt := range stmts {
		if len(stmt) > 0 {
			expressions = append(expressions, stmt)
		}
	}

	if len(expressions) == 0 {
		return "default_t()"
	}
	return strings.Join(expressions, " && ")
}

// hostStatement builds host statement in condition, host type is judged by the way
func hostStatement(host string) (string, int) {
	if len(host) == 0 {
		return "", HostTypeNoRestriction
	}

	if strings.HasPrefix(host, "*.") {
		return fmt.Sprintf(`req_host_suffix_in("%s")`, host[1:]), HostTypeWildcard
	} else {
		return fmt.Sprintf(`req_host_in("%s")`, host), HostTypeExact
	}
}

// hostStatement builds path statement in condition
// see: https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types
func pathStatement(path string, bfePathType networking.PathType) string {
	if len(path) == 0 {
		return "" // no restriction
	}

	if bfePathType == networking.PathTypeExact {
		return fmt.Sprintf(`req_path_in("%s", false)`, path)
	} else {
		path = strings.TrimRight(path, "/")
		if len(path) == 0 {
			return "" // no restriction
		}
		return fmt.Sprintf(`(req_path_in("%s", false) || req_path_prefix_in("%s/", false))`, path, path)
	}
}

func (i *ingressRawRuleInfo) GetPathType() networking.PathType {
	return BfePathType(i.PathType)
}

// GetPathPattern return path pattern according to path type
// Return:
// 		prefix: {/path}/*
//		exact: {/path}
func (i *ingressRawRuleInfo) GetPathPattern() string {
	switch i.GetPathType() {
	case networking.PathTypeExact:
		return i.Path
	default:
		return path.Join(i.Path, "*")
	}
}

func newAdvancedRuleCoverage() *advancedRuleCoverage {
	cov := new(advancedRuleCoverage)
	cov.HostPath = make(map[string]map[string]bool)
	return cov
}

// Cover records host & path pattern covered by advanced rule
func (c *advancedRuleCoverage) Cover(advancedRule ingressRawRuleInfo) {
	if _, ok := c.HostPath[advancedRule.Host]; !ok {
		c.HostPath[advancedRule.Host] = make(map[string]bool)
	}
	c.HostPath[advancedRule.Host][advancedRule.GetPathPattern()] = true
}

// IsCovered checks if a basic rule be overlapped with any known advanced rule
func (c *advancedRuleCoverage) IsCovered(basicRule ingressRawRuleInfo) bool {
	if !strings.HasPrefix(basicRule.Host, "*.") {
		// basic rule with exact host only overlapped with advanced rule with same exact host
		if _, ok := c.HostPath[basicRule.Host]; !ok {
			return false
		}
		return c.isPathOverlapped(basicRule, basicRule.Host)
	} else {
		// basic rule with wildcard host overlapped with any advanced rule with longer host (both exact & suffix)
		basicRuleHost := basicRule.Host[1:] // "*.bar.foo" ==> ".bar.foo"
		for advancedRuleHost := range c.HostPath {
			if strings.HasSuffix(advancedRuleHost, basicRuleHost) && c.isPathOverlapped(basicRule, advancedRuleHost) {
				return true
			}
		}
	}
	return false
}

// isPathOverlapped check if a basic rule be overlapped with any known advanced rule with given host
func (c *advancedRuleCoverage) isPathOverlapped(basicRule ingressRawRuleInfo, advancedRuleHost string) bool {
	switch basicRule.GetPathType() {
	case networking.PathTypeExact:
		// basic rule with exact path only overlapped with advanced rule with same exact path
		return c.HostPath[advancedRuleHost][basicRule.Path]
	default:
		// basic rule with prefix path overlapped with any advanced rule with longer path (both exact & prefix)
		for advancedRulePath := range c.HostPath[advancedRuleHost] {
			if strings.HasPrefix(advancedRulePath, basicRule.Path) {
				return true
			}
		}
		return false
	}
}
