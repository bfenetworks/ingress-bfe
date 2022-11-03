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

package rewrite

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "github.com/bfenetworks/bfe/bfe_basic/action"
	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/bfe/bfe_modules/mod_rewrite"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

const (
	ConfigNameRewrite = "mod_rewrite"
	RuleData          = "mod_rewrite/rewrite.data"
)

const (
	CallBackKey = "when"
	ParamKey    = "params"
	OrderKey    = "order"
)

const DefaultCallBackPoint = "AfterLocation"

type ModRewriteConfig struct {
	version          string
	rewriteRuleCache *rewriteRuleCache
	rewriteConfFile  *mod_rewrite.ReWriteConfFile
}

func NewRewriteConfig(version string) *ModRewriteConfig {
	return &ModRewriteConfig{
		version:          version,
		rewriteRuleCache: newRewriteRuleCache(version),
		rewriteConfFile:  newRewriteConfFile(version),
	}
}

func newRewriteConfFile(version string) *mod_rewrite.ReWriteConfFile {
	ruleFileList := make([]mod_rewrite.ReWriteRuleFile, 0)
	productRulesFile := make(mod_rewrite.ProductRulesFile)
	productRulesFile[configs.DefaultProduct] = (*mod_rewrite.RuleFileList)(&ruleFileList)
	return &mod_rewrite.ReWriteConfFile{
		Version: &version,
		Config:  &productRulesFile,
	}
}

func (c *ModRewriteConfig) Name() string {
	return ConfigNameRewrite
}

func (c *ModRewriteConfig) UpdateIngress(ingress *netv1.Ingress) error {
	// clear cache
	ingressName := util.NamespacedName(ingress.Namespace, ingress.Name)
	if c.rewriteRuleCache.ContainsIngress(ingressName) {
		c.rewriteRuleCache.DeleteByIngress(ingressName)
	}
	// nothing to update
	if len(ingress.Spec.Rules) == 0 {
		return nil
	}

	return c.rewriteRuleCache.UpdateByIngress(ingress)
}

func getRelatedAnnotationCnt(annots map[string]string, keys []string) int {
	cnt := 0
	for _, key := range keys {
		if _, ok := annots[key]; ok {
			cnt++
		}
	}
	return cnt
}

func checkAnnotationCnt(annots map[string]string) error {
	// check host annotation cnt
	hostAnnotations := []string{
		annotations.RewriteURLHostSetAnnotation,
		annotations.RewriteURLHostFromPathAnnotation,
	}
	if getRelatedAnnotationCnt(annots, hostAnnotations) == 2 {
		return fmt.Errorf("setting annotations %s and %s at the same time is not allowed", annotations.RewriteURLHostSetAnnotation, annotations.RewriteURLHostFromPathAnnotation)
	}

	// check path annotation cnt
	pathAnnotations := []string{
		annotations.RewriteURLPathSetAnnotation,
		annotations.RewriteURLPathPrefixAddAnnotation,
		annotations.RewriteURLPathPrefixTrimAnnotation,
	}
	if annots[annotations.RewriteURLPathSetAnnotation] != "" && getRelatedAnnotationCnt(annots, pathAnnotations) > 1 {
		return errors.New("when set a fixed url-path annotation, setting path-prefix-add or path-prefix-trim annotation is not allowed")
	}
	return nil
}

func parseCompositeParam(mp map[string]string) (string, string, int, error) {
	params := mp[ParamKey]
	if params == "" {
		return "", "", 0, errors.New("missing \"params\" field in rewrite-url action")
	}

	orderStr := mp[OrderKey]
	var order int64
	var err error
	if orderStr != "" {
		order, err = strconv.ParseInt(orderStr, 10, 64)
		if err != nil {
			return "", "", 0, err
		}
	}

	callback := mp[CallBackKey]
	if callback != "" {
		err = checkCallBackPoint(callback)
		if err != nil {
			return "", "", 0, err
		}
	} else {
		callback = DefaultCallBackPoint
	}
	return params, callback, int(order), nil
}

func getActionParamAndOrder(rewriteActions annotations.RewriteActions) (map[string]map[string][]string, map[string]map[string]int, error) {
	// actions: callback -> action -> params([]string)
	actions := make(map[string]map[string][]string, len(rewriteActions))
	// actionsOrder: callback -> action -> order
	actionsOrder := make(map[string]map[string]int)

	for cmd, cmdParam := range rewriteActions {
		for _, params := range cmdParam {
			actionParam, callback, order, e := parseCompositeParam(params)
			if e != nil {
				return nil, nil, e
			}

			if _, ok := actionsOrder[callback]; !ok {
				actions[callback] = make(map[string][]string)
				actionsOrder[callback] = make(map[string]int)
			}

			if _, ok := actionsOrder[callback][cmd]; ok {
				return nil, nil, errors.New("setting a rewrite-action with duplicate callback points is not allowed")
			}
			actionsOrder[callback][cmd] = order

			param, e := getActionParam(cmd, actionParam)
			if e != nil {
				return nil, nil, e
			}
			actions[callback][cmd] = param
		}
	}
	return actions, actionsOrder, nil
}

// parseRewriteAction parse ingress rewrite-url annotations and return rewrite actions and action order.
// the actions are divided by callback point.
func parseRewriteAction(annots map[string]string) (map[string]map[string][]string, map[string]map[string]int, error) {
	cntErr := checkAnnotationCnt(annots)
	if cntErr != nil {
		return nil, nil, cntErr
	}

	// convert annotation key & value to rewrite actions
	rawActions, parseErr := annotations.GetRewriteActions(annots)
	if parseErr != nil {
		return nil, nil, parseErr
	}

	// parse rewrite actions
	return getActionParamAndOrder(rawActions)
}

func checkCallBackPoint(cb string) error {
	switch cb {
	case DefaultCallBackPoint:
		return nil
	default:
		return fmt.Errorf("%s callback point in rewrite-url action is not supported", cb)
	}
}

func checkHost(host string) error {
	// wildcard hostname: started with "*." is allowed
	if strings.Count(host, "*") > 1 || (strings.Count(host, "*") == 1 && !strings.HasPrefix(host, "*.")) {
		return fmt.Errorf("wildcard host[%s] is illegal, should start with *. ", host)
	}
	return nil
}

func checkPath(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("path is not set")
	}

	if strings.ContainsAny(path, "*") {
		return fmt.Errorf("path[%s] is illegal", path)
	}
	return nil
}

func getActionParam(cmd, param string) ([]string, error) {
	switch cmd {
	case ActionHostSet:
		if err := checkHost(param); err != nil {
			return nil, fmt.Errorf("invalid host name, error: %s", err)
		}
		return []string{param}, nil
	case ActionHostSetFromPathPrefix:
		if param == "" || strings.ToLower(param) != "true" || strings.ToLower(param) != "t" {
			return nil, fmt.Errorf("invalid host-set-from-path param: %s, which should be true or t", param)
		}
		return []string{}, nil
	case ActionPathSet, ActionPathPrefixAdd, ActionPathPrefixTrim, ActionQueryDelAllExcept:
		if err := checkPath(param); err != nil {
			return nil, err
		}
		return []string{param}, nil
	case ActionQueryAdd, ActionQueryRename:
		paramMaps := make(map[string]string)
		err := json.Unmarshal([]byte(param), &paramMaps)
		if err != nil {
			return nil, err
		}
		if len(paramMaps) == 0 {
			return nil, fmt.Errorf("the param of annotation %s can not be empty", cmd)
		}
		paramList := make([]string, 0)
		for k, v := range paramMaps {
			paramList = append(paramList, k, v)
		}
		return paramList, nil
	case ActionQueryDel:
		paramList := make([]string, 0)
		err := json.Unmarshal([]byte(param), &paramList)
		if err != nil {
			return nil, err
		}
		if len(paramList) == 0 {
			return nil, fmt.Errorf("the param of annotation %s can not be empty", cmd)
		}
		return paramList, nil
	default:
		return nil, fmt.Errorf("unsupported annotation for rewrite action: %s", cmd)
	}
}

func (c *ModRewriteConfig) DeleteIngress(namespace, name string) {
	ingressName := util.NamespacedName(namespace, name)
	if !c.rewriteRuleCache.ContainsIngress(ingressName) {
		return
	}

	c.rewriteRuleCache.DeleteByIngress(ingressName)
}

func (c *ModRewriteConfig) Reload() error {
	if err := c.updateRewriteConf(); err != nil {
		return fmt.Errorf("update %s config error: %v", RuleData, err)
	}

	if *c.rewriteConfFile.Version != c.version {
		// dump config file
		err := util.DumpBfeConf(RuleData, c.rewriteConfFile)
		if err != nil {
			return fmt.Errorf("dump %s error: %v", RuleData, err)
		}
		// reload bfe engine
		err = util.ReloadBfe(ConfigNameRewrite)
		if err != nil {
			return err
		}
		c.version = *c.rewriteConfFile.Version
	}

	return nil
}

func getRewriteConfSegment(callback string) (string, error) {
	switch callback {
	case DefaultCallBackPoint:
		return configs.DefaultProduct, nil
	default:
		return "", fmt.Errorf("setting rewrite-action at %s callback point is not supported now", callback)
	}
}

func (c *ModRewriteConfig) updateRewriteConf() error {
	if *c.rewriteConfFile.Version == c.rewriteRuleCache.Version {
		return nil
	}

	ruleList := c.rewriteRuleCache.GetRules()
	segmentRules := make(map[string]mod_rewrite.RuleFileList)
	for _, rule := range ruleList {
		rule := rule.(*rewriteRule)
		cond, err := rule.GetCond()
		if err != nil {
			return err
		}
		if _, ok := segmentRules[rule.when]; !ok {
			segmentRules[rule.when] = make(mod_rewrite.RuleFileList, 0, len(ruleList))
		}
		segmentRules[rule.when] = append(segmentRules[rule.when], mod_rewrite.ReWriteRuleFile{
			Cond:    &cond,
			Actions: rule.actions,
			Last:    rule.last,
		})
	}

	rewriteConfFile := newRewriteConfFile(c.rewriteRuleCache.Version)
	// map rule to config segment through callback point
	for cb, rules := range segmentRules {
		seg, err := getRewriteConfSegment(cb)
		if err != nil {
			return nil
		}
		(*rewriteConfFile.Config)[seg] = &rules
	}

	if err := mod_rewrite.ReWriteConfCheck(*rewriteConfFile); err != nil {
		return nil
	}

	c.rewriteConfFile = rewriteConfFile
	return nil
}
