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
	"fmt"
	"strings"
	"time"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
)

type BaseRule struct {
	Ingress     string
	Host        string
	Path        string
	Annotations map[string]string
	CreateTime  time.Time
}

func (rule BaseRule) GetIngress() string {
	return rule.Ingress
}

func (rule BaseRule) GetHost() string {
	return rule.Host
}

func (rule BaseRule) GetPath() string {
	return rule.Path
}

func (rule BaseRule) GetAnnotations() map[string]string {
	return rule.Annotations
}

func (rule BaseRule) GetCreateTime() time.Time {
	return rule.CreateTime
}

func (rule BaseRule) GetCond() (string, error) {
	return buildCondition(rule.Host, rule.Path, rule.Annotations)
}

func buildCondition(host string, path string, annots map[string]string) (string, error) {
	var statement []string

	primitive, err := hostPrimitive(host)
	if err != nil {
		return "", err
	}
	if len(primitive) > 0 {
		statement = append(statement, primitive)
	}

	primitive, err = pathPrimitive(path)
	if err != nil {
		return "", err
	}
	if len(primitive) > 0 {
		statement = append(statement, primitive)
	}

	primitive, err = annotations.GetRouteExpression(annots)
	if err != nil {
		return "", err
	}
	if len(primitive) > 0 {
		statement = append(statement, primitive)
	}

	return strings.Join(statement, "&&"), nil
}

// hostPrimitive builds host primitive in condition
func hostPrimitive(host string) (string, error) {
	if len(host) == 0 || host == "*" {
		return "", nil
	}

	if strings.HasPrefix(host, "*.") {
		dn := host[2:]
		dn = strings.ReplaceAll(dn, ".", "\\.")
		return fmt.Sprintf(`req_host_regmatch("(?i)^[^\.]+%s")`, dn), nil
	}
	return fmt.Sprintf(`req_host_in("%s")`, host), nil
}

// pathPrimitive builds path primitive in condition
func pathPrimitive(path string) (string, error) {
	if len(path) == 0 || path == "*" {
		return "", nil // no restriction
	}
	if path[len(path)-1] == '*' {
		return fmt.Sprintf(`req_path_element_prefix_in("%s", false)`, path[:len(path)-1]), nil
	}
	return fmt.Sprintf(`req_path_in("%s", false)`, path), nil
}
