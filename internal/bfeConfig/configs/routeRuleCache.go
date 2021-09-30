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
	"sort"
	"strings"
	"time"

	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/jwangsadinata/go-multimap/setmultimap"
)

type httpRule struct {
	ingress     string
	host        string
	path        string
	annotations map[string]string
	cluster     string
	createTime  time.Time
}

type HttpRouteRuleCache struct {
	// ingress -> rules
	ingress2Rule *setmultimap.MultiMap

	// host -> path -> rule
	ruleMap map[string]map[string][]*httpRule
}

type RouteRuleCache struct {
	httpRules HttpRouteRuleCache
}

func NewRouteRuleCache() *RouteRuleCache {
	return &RouteRuleCache{
		HttpRouteRuleCache{
			ingress2Rule: setmultimap.New(),
			ruleMap:      make(map[string]map[string][]*httpRule),
		},
	}
}

func NewHttpRule(ingress string, host string, path string, annots map[string]string, cluster string, time time.Time) *httpRule {
	return &httpRule{
		ingress:     ingress,
		host:        host,
		path:        path,
		annotations: annots,
		cluster:     cluster,
		createTime:  time,
	}
}

func (c *RouteRuleCache) GetHttpRules() (basicRuleList []*httpRule, advancedRuleList []*httpRule) {
	return c.httpRules.get()
}

func (c *RouteRuleCache) PutHttpRule(rule *httpRule) error {
	return c.httpRules.put(rule)
}

func (c *RouteRuleCache) DeleteHttpRulesByIngress(ingress string) {
	c.httpRules.delete(ingress)
}

// ContainsIngress returns true if ingress exist in cache
func (c *RouteRuleCache) ContainsIngress(ingress string) bool {
	return c.httpRules.ingress2Rule.ContainsKey(ingress)
}

func (c *HttpRouteRuleCache) get() (basicRuleList []*httpRule, advancedRuleList []*httpRule) {
	for _, paths := range c.ruleMap {
		for _, rules := range paths {
			if len(rules) == 0 {
				continue
			}

			// add host+path rule to basic rule list
			if len(rules) == 1 && annotations.Priority(rules[0].annotations) == annotations.PriorityBasic {
				basicRuleList = append(basicRuleList, rules[0])
				continue
			}
			// add a fake basicRule,cluster=ADVANCED_MODE
			newRule := *rules[0]
			newRule.cluster = route_rule_conf.AdvancedMode
			basicRuleList = append(basicRuleList, &newRule)

			// add advanced rule
			advancedRuleList = append(advancedRuleList, rules...)
		}
	}

	// host: exact match over wildcard match
	// path: long path over short path
	sort.SliceStable(advancedRuleList, func(i, j int) bool {
		// compare host
		if result := comparePriority(advancedRuleList[i].host, advancedRuleList[j].host, wildcardHost); result != 0 {
			return result > 0
		}

		// compare path
		if result := comparePriority(advancedRuleList[i].path, advancedRuleList[j].path, wildcardPath); result != 0 {
			return result > 0
		}

		// compare annotation
		priority1 := annotations.Priority(advancedRuleList[i].annotations)
		priority2 := annotations.Priority(advancedRuleList[j].annotations)
		if priority1 != priority2 {
			return priority1 > priority2
		}

		// check createTime
		return advancedRuleList[i].createTime.Before(advancedRuleList[j].createTime)
	})

	return
}

func (c *HttpRouteRuleCache) delete(ingressName string) {
	deleteRules, _ := c.ingress2Rule.Get(ingressName)

	// delete rules from ruleMap
	for _, rule := range deleteRules {
		rule := rule.(*httpRule)
		rules, ok := c.ruleMap[rule.host][rule.path]
		if !ok {
			continue
		}
		c.ruleMap[rule.host][rule.path] = delRule(rules, ingressName)
		if len(c.ruleMap[rule.host][rule.path]) == 0 {
			delete(c.ruleMap[rule.host], rule.path)
		}
		if len(c.ruleMap[rule.host]) == 0 {
			delete(c.ruleMap, rule.host)
		}
	}

	c.ingress2Rule.RemoveAll(ingressName)
}

func (c *HttpRouteRuleCache) put(rule *httpRule) error {
	if _, ok := c.ruleMap[rule.host]; !ok {
		c.ruleMap[rule.host] = make(map[string][]*httpRule)
	}

	for i, r := range c.ruleMap[rule.host][rule.path] {
		if annotations.Equal(rule.annotations, r.annotations) {
			// all conditions are same, oldest rule is valid
			if rule.createTime.Before(r.createTime) {
				log.V(0).Info("rule is overwritten by elder ingress", "ingress", r.ingress, "host", r.host, "path", r.path, "old-ingress", rule.ingress)

				c.ingress2Rule.Remove(rule.ingress, c.ruleMap[rule.host][rule.path][i])
				c.ruleMap[rule.host][rule.path][i] = rule
				c.ingress2Rule.Put(rule.ingress, rule)
				return nil
			} else if rule.createTime.Equal(r.createTime) {
				return nil
			} else {
				return fmt.Errorf("conflict with %s, rule [host: %s, path: %s]", r.ingress, rule.host, rule.path)
			}
		}
	}
	c.ingress2Rule.Put(rule.ingress, rule)
	c.ruleMap[rule.host][rule.path] = append(c.ruleMap[rule.host][rule.path], rule)

	return nil
}

func delRule(ruleList []*httpRule, ingress string) []*httpRule {
	var result []*httpRule
	for _, rule := range ruleList {
		if rule.ingress != ingress {
			result = append(result, rule)
		}
	}
	return result
}

func comparePriority(str1, str2 string, wildcard func(string) bool) int {
	// non-wildcard has higher priority
	if !wildcard(str1) && wildcard(str2) {
		return 1
	}
	if wildcard(str1) && !wildcard(str2) {
		return -1
	}

	// longer host has higher priority
	if len(str1) > len(str2) {
		return 1
	} else if len(str1) == len(str2) {
		return 0
	} else {
		return -1
	}

}

func wildcardPath(path string) bool {
	if len(path) > 0 && strings.HasSuffix(path, "*") {
		return true
	}

	return false
}

func wildcardHost(host string) bool {
	if len(host) > 0 && strings.HasPrefix(host, "*.") {
		return true
	}

	return false
}
