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

package annotations

import (
	"encoding/json"
	"fmt"

	. "github.com/bfenetworks/bfe/bfe_basic/action"
)

const rewriteAnnotationPrefix = BfeAnnotationPrefix + "rewrite-url."

// The annotation related to rewrite host action.
const (
	RewriteURLHostSetAnnotation      = rewriteAnnotationPrefix + "host"
	RewriteURLHostFromPathAnnotation = rewriteAnnotationPrefix + "host-from-path"
)

// The annotation related to rewrite path action.
const (
	RewriteURLPathSetAnnotation        = rewriteAnnotationPrefix + "path"
	RewriteURLPathPrefixAddAnnotation  = rewriteAnnotationPrefix + "path-prefix-add"
	RewriteURLPathPrefixTrimAnnotation = rewriteAnnotationPrefix + "path-prefix-trim"
)

// The annotation related to rewrite query action.
const (
	RewriteURLQueryAddAnnotation             = rewriteAnnotationPrefix + "query-add"
	RewriteURLQueryDeleteAnnotation          = rewriteAnnotationPrefix + "query-delete"
	RewriteURLQueryRenameAnnotation          = rewriteAnnotationPrefix + "query-rename"
	RewriteURLQueryDeleteAllExceptAnnotation = rewriteAnnotationPrefix + "query-delete-all-except"
)

// transRewriteAnnotationKeyToAction convert bfe-ingress rewrite annotation key to BFE engine rewrite action key
func transRewriteAnnotationKeyToAction(annot string) (ac string) {
	switch annot {
	case RewriteURLHostSetAnnotation:
		ac = ActionHostSet
	case RewriteURLHostFromPathAnnotation:
		ac = ActionHostSetFromPathPrefix
	case RewriteURLPathSetAnnotation:
		ac = ActionPathSet
	case RewriteURLPathPrefixAddAnnotation:
		ac = ActionPathPrefixAdd
	case RewriteURLPathPrefixTrimAnnotation:
		ac = ActionPathPrefixTrim
	case RewriteURLQueryAddAnnotation:
		ac = ActionQueryAdd
	case RewriteURLQueryDeleteAnnotation:
		ac = ActionQueryDel
	case RewriteURLQueryRenameAnnotation:
		ac = ActionQueryRename
	case RewriteURLQueryDeleteAllExceptAnnotation:
		ac = ActionQueryDelAllExcept
	default:
		ac = ""
	}
	return
}

// RewriteActions define struct of rewrite annotation and params.
// examples: {"HOST_SET": [{"params": baidu.com, "when": "AfterLocation", "order": "1"}]}.
// When BFE engine add new rewrite action callback points,
// user can set "when" field to other callback points
// and add "cond" field for condition setting.
type RewriteActions map[string]RewriteActionParam
type RewriteActionParam []map[string]string

// GetRewriteActions try to parse the cmd and the param of the rewrite action from the annotations
func GetRewriteActions(annotations map[string]string) (RewriteActions, error) {
	rewriteActions := make(RewriteActions, 0)
	for annot, v := range annotations {
		if cmd := transRewriteAnnotationKeyToAction(annot); cmd != "" {
			param := make(RewriteActionParam, 0)
			err := json.Unmarshal([]byte(v), &param)
			if err != nil {
				return nil, fmt.Errorf("annotation %s's param is illegal, error: %s", annot, err)
			}
			rewriteActions[cmd] = param
		}
	}
	return rewriteActions, nil
}
