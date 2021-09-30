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
package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bfenetworks/ingress-bfe/internal/option"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
)

func ClusterName(ingressName string, backend *netv1.IngressServiceBackend) string {
	port := backend.Port.Name
	if backend.Port.Number > 0 {
		port = fmt.Sprintf("%d", backend.Port.Number)
	}
	return fmt.Sprintf("%s_%s_%s", ingressName, backend.Name, port)
}

// DefaultClusterName returns a default cluster for default backend
func DefaultClusterName() string {
	ingress := "__defaultCluster__"
	return fmt.Sprintf("%s_%s_%d", ingress, option.Opts.DefaultBackend, 0)
}

func ParsePort(clusterName string) netv1.ServiceBackendPort {
	port := netv1.ServiceBackendPort{}
	index := strings.LastIndexByte(clusterName, '_')
	if index < 0 {
		return port
	}
	portStr := clusterName[index+1:]

	if i, err := strconv.Atoi(portStr); err == nil {
		port.Number = int32(i)
	} else {
		port.Name = portStr
	}
	return port
}

func NamespacedName(namespace, name string) string {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}.String()
}

func SplitNamespacedName(namespacedName string) (namespace, name string) {
	names := strings.Split(namespacedName, "/")
	if len(names) != 2 {
		return "", ""
	}
	return names[0], names[1]
}
