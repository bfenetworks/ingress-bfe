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
	"reflect"
)

var (
	lastConfigMap map[string]interface{}
)

func isConfigEqual(configName string, newConfig interface{}) bool {
	lastConfig, ok := lastConfigMap[configName]
	if !ok {
		return false
	}

	return reflect.DeepEqual(lastConfig, newConfig)
}

func updateLastConfig(configName string, newConfig interface{}) {
	lastConfigMap[configName] = newConfig
}
