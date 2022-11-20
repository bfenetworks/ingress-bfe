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

//Package rewrite is the module of rewrite url.
//This file implements operate rule cache, generate and reload config file methods.
package rewrite

import (
	"fmt"

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
	for cb := range segmentRules {
		err := annotations.CheckCallBackPoint(cb)
		if err != nil {
			return err
		}
	}
	afterLocationRules := segmentRules[annotations.DefaultCallBackPoint]
	(*rewriteConfFile.Config)[configs.DefaultProduct] = &afterLocationRules

	if err := mod_rewrite.ReWriteConfCheck(*rewriteConfFile); err != nil {
		return err
	}

	c.rewriteConfFile = rewriteConfFile
	return nil
}
