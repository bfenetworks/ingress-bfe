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
	"sort"
	"time"

	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/cache"
)

type routeRule struct {
	Cluster string
	*cache.BaseRule
}

type RouteRuleCache struct {
	*cache.BaseCache
}

func newRouteRuleCache(version string) *RouteRuleCache {
	return &RouteRuleCache{
		BaseCache: cache.NewBaseCache(version),
	}
}

func newRouteRule(ingress string, host string, path string, annots map[string]string, cluster string, time time.Time) *routeRule {
	return &routeRule{
		BaseRule: cache.NewBaseRule(
			ingress,
			host,
			path,
			annots,
			time,
		),
		Cluster: cluster,
	}
}

func (c *RouteRuleCache) getRouteRules() (basicRuleList []*routeRule, advancedRuleList []*routeRule) {
	httpRules := c.BaseRules
	for _, paths := range httpRules.RuleMap {
		for _, ruleList := range paths {
			if len(ruleList) == 0 {
				continue
			}

			// add host+path rule to basic rule list
			if len(ruleList) == 1 && annotations.Priority(ruleList[0].GetAnnotations()) == annotations.PriorityBasic {
				basicRuleList = append(basicRuleList, ruleList[0].(*routeRule))
				continue
			}
			// add a fake basicRule,cluster=ADVANCED_MODE
			newRule := *ruleList[0].(*routeRule)
			newRule.Cluster = route_rule_conf.AdvancedMode
			basicRuleList = append(basicRuleList, &newRule)

			// add advanced rule
			for _, rule := range ruleList {
				advancedRuleList = append(advancedRuleList, rule.(*routeRule))
			}
		}
	}

	// host: exact match over wildcard match
	// path: long path over short path
	sort.SliceStable(advancedRuleList, func(i, j int) bool {
		return cache.CompareRule(advancedRuleList[i], advancedRuleList[j])
	})

	return
}

func (c *RouteRuleCache) UpdateByIngress(ingress *netv1.Ingress) error {
	return c.BaseCache.UpdateByIngressFramework(
		ingress,
		nil,
		nil,
		func(ingress *netv1.Ingress, host, path string, httpPath netv1.HTTPIngressPath) (cache.Rule, error) {
			ingressName := util.NamespacedName(ingress.Namespace, ingress.Name)
			return newRouteRule(
				ingressName,
				host,
				path,
				ingress.Annotations,
				util.ClusterName(ingressName, httpPath.Backend.Service),
				ingress.CreationTimestamp.Time,
			), nil
		},
	)
}
