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
)

const (
	WeightKey        = "balance.weight"
	WeightAnnotation = BfeAnnotationPrefix + WeightKey
)

// ServicesWeight define struct of annotation "balance.weight"
// example: {"service": {"service1":80, "service2":20}}
type ServicesWeight map[string]int
type Balance map[string]ServicesWeight

// GetBalance parse annotation "balance.weight"
func GetBalance(annotations map[string]string) (Balance, error) {
	value, ok := annotations[WeightAnnotation]
	if !ok {
		return nil, nil
	}

	var lb = make(Balance)
	err := json.Unmarshal([]byte(value), &lb)
	if err != nil {
		return nil, fmt.Errorf("annotation %s is illegal, error: %s", WeightAnnotation, err)
	}

	// check whether weight sum > 0
	for _, services := range lb {
		sum := 0
		for _, weight := range services {
			if weight < 0 {
				return nil, fmt.Errorf("weight of load balance service should >= 0")
			}
			sum += weight
		}
		if sum == 0 {
			return nil, fmt.Errorf("sum of all load balance service weight should > 0")
		}
	}
	return lb, nil
}
