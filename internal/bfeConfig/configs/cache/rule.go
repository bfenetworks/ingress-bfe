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

package cache

import (
	"strings"
	"time"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
)

// Rule is an abstraction of BFE Rule.
// The BFE Rule is the basis for BFE Engine to process a Request.
// For example, the route module need to use route rules to detect which
// backend service to route a Request to;
// the redirect module need to use redirect rules to decide
// if a Request should be redirected, and how.
// The route rule and redirect rule are both Rules.
type Rule interface {
	// GetIngress gets the namespaced name of the ingress to which the Rule belongs
	GetIngress() string

	// GetHost gets the host of the Rule
	GetHost() string

	// GetPath gets the path of the Rule
	GetPath() string

	// GetAnnotations gets the annotations of the ingress to which the Rule belongs
	GetAnnotations() map[string]string

	// GetCreateTime gets the created time of the Rule
	GetCreateTime() time.Time

	// GetCond generates a BFE Condition from the Rule
	GetCond() (string, error)
}

// CompareRule compares the priority of two Rules.
// The function can be used to sort a Rule list.
func CompareRule(rule1, rule2 Rule) bool {
	// host: exact match over wildcard match
	// path: long path over short path

	// compare host
	if result := comparePriority(rule1.GetHost(), rule2.GetHost(), wildcardHost); result != 0 {
		return result > 0
	}

	// compare path
	if result := comparePriority(rule1.GetPath(), rule2.GetPath(), wildcardPath); result != 0 {
		return result > 0
	}

	// compare annotation
	priority1 := annotations.Priority(rule1.GetAnnotations())
	priority2 := annotations.Priority(rule2.GetAnnotations())
	if priority1 != priority2 {
		return priority1 > priority2
	}

	// check createTime
	return rule1.GetCreateTime().Before(rule2.GetCreateTime())
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
