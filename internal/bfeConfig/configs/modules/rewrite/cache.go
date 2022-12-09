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
//This file defines rewrite & cache's struct, also implements update ingress method.
package rewrite

import (
	"sort"

	netv1 "k8s.io/api/networking/v1"

	bfeac "github.com/bfenetworks/bfe/bfe_basic/action"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/cache"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

type rewriteRule struct {
	*cache.BaseRule
	actions []bfeac.Action
	//last means whether to continue match other conditions, only support true in this version
	last *bool
	// when defines callback point, only support AfterLocation in this version
	when string // default AfterLocation
}

type rewriteRuleCache struct {
	*cache.BaseCache
}

func newRewriteRuleCache(version string) *rewriteRuleCache {
	return &rewriteRuleCache{
		BaseCache: cache.NewBaseCache(version),
	}
}

func (c rewriteRuleCache) UpdateByIngress(ingress *netv1.Ingress) error {
	rewriteActions, err := annotations.GetRewriteAction(ingress.Annotations)
	if err != nil {
		return err
	}

	if rewriteActions == nil {
		return nil
	}

	for callback, callbackActions := range rewriteActions {
		e := c.UpdateByIngressFramework(
			ingress,
			func(ingress *netv1.Ingress, host, path string, _ netv1.HTTPIngressPath) (cache.Rule, error) {
				actions := make([]bfeac.Action, 0)
				for cmd, p := range callbackActions {
					if cmd == "QUERY_ADD" || cmd == "QUERY_RENAME" {
						for i := 0; i < len(p.Params); i += 2 {
							ac := bfeac.Action{
								Cmd:    cmd,
								Params: p.Params[i : i+2],
							}
							actions = append(actions, ac)
						}
					} else if cmd == "PATH_STRIP" {
						prefix, err := annotations.GetPathStripPrefix(path, p.Params[0])
						if err != nil {
							return nil, err
						}
						ac := bfeac.Action{
							Cmd:    "PATH_PREFIX_TRIM",
							Params: []string{prefix},
						}
						actions = append(actions, ac)
					} else {
						ac := bfeac.Action{
							Cmd:    cmd,
							Params: p.Params,
						}
						actions = append(actions, ac)
					}
				}
				// sort by user defined action order
				sort.SliceStable(actions, func(i, j int) bool {
					return callbackActions[actions[i].Cmd].Order < callbackActions[actions[j].Cmd].Order
				})
				last := true
				return &rewriteRule{
					BaseRule: cache.NewBaseRule(
						util.NamespacedName(ingress.Namespace, ingress.Name),
						host,
						path,
						ingress.Annotations,
						ingress.CreationTimestamp.Time,
					),
					actions: actions,
					when:    callback,
					last:    &last,
				}, nil
			},
			nil,
			nil,
		)
		if e != nil {
			return e
		}
	}
	return nil
}
