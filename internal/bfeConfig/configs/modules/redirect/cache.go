// Copyright (c) 2022 The BFE Authors.
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

package redirect

import (
	"github.com/bfenetworks/bfe/bfe_modules/mod_redirect"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	netv1 "k8s.io/api/networking/v1"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/cache"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
)

type redirectRule struct {
	*cache.BaseRule
	// statusCode is the response status code of redirect
	statusCode int
	// action is the redirect action. Refer to https://www.bfe-networks.net/en_us/modules/mod_redirect/mod_redirect/.
	action *mod_redirect.ActionFileList
}

type redirectRuleCache struct {
	*cache.BaseCache
}

func newRedirectRuleCache(version string) *redirectRuleCache {
	return &redirectRuleCache{
		BaseCache: cache.NewBaseCache(version),
	}
}

func (c redirectRuleCache) UpdateByIngress(ingress *netv1.Ingress) error {
	if len(ingress.Spec.Rules) == 0 {
		return nil
	}

	cmd, param, err := parseRedirectActionFromAnnotations(ingress.Annotations)
	if err != nil {
		return err
	}
	statusCode, err := annotations.GetRedirectStatusCode(ingress.Annotations)
	if err != nil {
		return err
	}
	return c.BaseCache.UpdateByIngressFramework(
		ingress,
		func(ingress *netv1.Ingress, host, path string, _ netv1.HTTPIngressPath) (cache.Rule, error) {
			// preCheck
			if err := checkAction(cmd, param); err != nil {
				return nil, err
			}
			if err := checkStatusCode(statusCode); err != nil {
				return nil, err
			}

			action := &mod_redirect.ActionFileList{mod_redirect.ActionFile{
				Cmd:    &cmd,
				Params: []string{param},
			}}
			if err = mod_redirect.ActionFileListCheck(action); err != nil {
				return nil, err
			}
			return &redirectRule{
				BaseRule: cache.NewBaseRule(
					util.NamespacedName(ingress.Namespace, ingress.Name),
					host,
					path,
					ingress.Annotations,
					ingress.CreationTimestamp.Time,
				),
				statusCode: statusCode,
				action:     action,
			}, nil
		},
		func() (bool, error) {
			return cmd != "", err
		},
		nil,
	)
}
