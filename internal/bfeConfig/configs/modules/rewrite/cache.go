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
	"sort"

	netv1 "k8s.io/api/networking/v1"

	. "github.com/bfenetworks/bfe/bfe_basic/action"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/cache"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

type rewriteRule struct {
	*cache.BaseRule
	actions []Action
	//last means whether to continue match other conditions, only support true in this version
	last *bool
	//cond is rewrite action's condition, but not supported in this version
	cond map[string]string
	// when defines callback point, only support AfterLocation in this version
	when string // default and only support AfterLocation
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
	parsedActions, actionOrder, err := parseRewriteAction(ingress.Annotations)
	if err != nil {
		return err
	}

	if len(parsedActions) == 0 {
		return nil
	}

	for callback, callbackParams := range parsedActions {
		e := c.UpdateByIngressFramework(
			ingress,
			func(ingress *netv1.Ingress, host, path string, _ netv1.HTTPIngressPath) (cache.Rule, error) {
				actions := make([]Action, 0)
				for cmd, params := range callbackParams {
					if cmd == ActionQueryAdd || cmd == ActionQueryRename {
						for i := 0; i < len(params); i += 2 {
							ac := Action{
								Cmd:    cmd,
								Params: params[i : i+2],
							}
							actions = append(actions, ac)
						}
					} else {
						ac := Action{
							Cmd:    cmd,
							Params: params,
						}
						actions = append(actions, ac)
					}
				}
				// sort by user defined action order
				sort.SliceStable(actions, func(i, j int) bool {
					return actionOrder[callback][actions[i].Cmd] < actionOrder[callback][actions[j].Cmd]
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
