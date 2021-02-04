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
package bfe_ingress

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	BfeAnnotationPrefix = "bfe.ingress.kubernetes.io/"
	CookieKey           = "router.cookie"
	HeaderKey           = "router.header"

	LoadBalanceKey        = "loadbalance"
	LoadBalanceAnnotation = BfeAnnotationPrefix + LoadBalanceKey

	CookiePriority = 0
	HeaderPriority = 1
)

type BfeAnnotation interface {
	Priority() int
	Check() error
	Build() string
}

type cookieAnnotation struct {
	annotationStr string
}

func (cookie *cookieAnnotation) Priority() int {
	return CookiePriority
}

// cookieAnnotation.Check; cookie annotation's valid str is "key: value"
func (cookie *cookieAnnotation) Check() error {
	if len(strings.Split(cookie.annotationStr, ":")) < 2 {
		return fmt.Errorf("invalid cookie annotation str")
	}
	return nil
}

func (cookie *cookieAnnotation) Build() string {
	strs := strings.Split(cookie.annotationStr, ":")
	key := strs[0]
	key = strings.TrimSpace(key)
	value := strings.Join(strs[1:], ":")
	value = strings.TrimSpace(value)
	con := fmt.Sprintf("req_cookie_value_in(\"%s\", \"%v\", false)", key, value)
	return con
}

type headerAnnotation struct {
	annotationStr string
}

func (header *headerAnnotation) Priority() int {
	return HeaderPriority
}

// headerAnnotation.Check; cookie annotation's valid str is "key: value"
func (header *headerAnnotation) Check() error {
	if len(strings.Split(header.annotationStr, ":")) < 2 {
		return fmt.Errorf("invalid header annotation str")
	}
	return nil
}

func (header *headerAnnotation) Build() string {
	strs := strings.Split(header.annotationStr, ":")
	key := strs[0]
	key = strings.TrimSpace(key)
	value := strings.Join(strs[1:], ":")
	value = strings.TrimSpace(value)
	con := fmt.Sprintf("req_header_value_in(\"%s\", \"%v\", false)", key, value)
	return con
}

func BuildBfeAnnotation(key string, value string) (BfeAnnotation, error) {
	if !strings.HasPrefix(key, BfeAnnotationPrefix) {
		return nil, fmt.Errorf("Unsupported annotation: %s", key)
	}
	newKey := strings.ReplaceAll(key, BfeAnnotationPrefix, "")
	var annotation BfeAnnotation
	switch newKey {
	case CookieKey:
		annotation = &cookieAnnotation{annotationStr: value}
	case HeaderKey:
		annotation = &headerAnnotation{annotationStr: value}
	default:
		return nil, fmt.Errorf("Unsupported annotation: %s", newKey)
	}
	if err := annotation.Check(); err != nil {
		return nil, err
	}
	return annotation, nil
}

func SortAnnotations(annotationConds []BfeAnnotation) {
	sort.Slice(annotationConds, func(i, j int) bool {
		return annotationConds[i].Priority() < annotationConds[j].Priority()
	})
}

type ServicesWeight map[string]int
type LoadBalance map[string]ServicesWeight

func (l *LoadBalance) ContainService(service string) bool {
	if l == nil || (*l) == nil {
		return false
	}
	_, ok := (*l)[service]
	return ok
}

func (l *LoadBalance) GetService(serviceName string) (ServicesWeight, error) {
	if !l.ContainService(serviceName) {
		return nil, fmt.Errorf("load balance donot contain[%s]", serviceName)
	}
	return (*l)[serviceName], nil
}

func BuildLoadBalanceAnnotation(key string, value string) (LoadBalance, error) {
	if key != LoadBalanceAnnotation {
		return nil, fmt.Errorf("Unsupported annotation: %s", key)
	}
	var lb = make(LoadBalance)
	err := json.Unmarshal([]byte(value), &lb)
	if err != nil {
		return nil, err
	}

	for _, services := range lb {
		sum := 0
		var tmpList = make([]string, 0)
		for name, weight := range services {
			if weight < 0 {
				return nil, fmt.Errorf("weight of load balance service should greate than or equal to zero")
			}
			sum += weight
			tmpList = append(tmpList, name)
		}
		//add sort make it easy to do unit test
		sort.Slice(tmpList, func(i, j int) bool {
			return tmpList[i] < tmpList[j]
		})
		if sum == 0 {
			return nil, fmt.Errorf("sum of load balance service weight is zero")
		}
		if sum != 100 {
			curSum := 0
			for index, name := range tmpList {
				weight := services[name]
				newWeight := int((float32(weight)/float32(sum)*100.0 + 0.5))
				if index == len(tmpList)-1 {
					newWeight = 100 - curSum
				}
				curSum += newWeight
				services[name] = newWeight
			}
		}
	}
	return lb, nil
}
