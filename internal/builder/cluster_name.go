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
package builder

import (
	"fmt"
	"strings"
)

import (
	networking "k8s.io/api/networking/v1beta1"
)

func ClusterName(ingress *networking.Ingress, balance LoadBalance, p networking.HTTPIngressPath) string {
	if !balance.ContainService(p.Backend.ServiceName) {
		return SingleClusterName(ingress.Namespace, p.Backend.ServiceName)
	}

	return MultiClusterName(ingress.Namespace, ingress.Name, p.Backend.ServiceName)
}

// SingleClusterName return cluster name for single k8s service
// e.g. "default_whoAmI"
func SingleClusterName(namespace, serviceName string) string {
	return fmt.Sprintf("%s_%s", namespace, serviceName)
}

// MultiClusterName return cluster name for multi k8s service
// e.g. "default_ingressTest_whoAmI"
func MultiClusterName(namespace, ingressName, serviceKey string) string {
	return fmt.Sprintf("%s_%s_%s", namespace, ingressName, serviceKey)
}

// Namespace return namespace which parsed from cluster name
func Namespace(clusterName string) string {
	return strings.Split(clusterName, "_")[0]
}
