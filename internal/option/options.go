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

package option

import (
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/bfenetworks/ingress-bfe/internal/option/ingress"
)

const (
	ClusterName            = "default"
	MetricsBindAddress     = ":9080"
	HealthProbeBindAddress = ":9081"
)

type Options struct {
	ClusterName string

	Namespaces      string
	NamespaceList   []string
	MetricsAddr     string
	HealthProbeAddr string

	Ingress *ingress.Options
}

var (
	Opts *Options
)

func NewOptions() *Options {
	return &Options{
		ClusterName:     ClusterName,
		Namespaces:      corev1.NamespaceAll,
		MetricsAddr:     MetricsBindAddress,
		HealthProbeAddr: HealthProbeBindAddress,
		Ingress:         ingress.NewOptions(),
	}
}

func SetOptions(option *Options) error {
	if err := option.Ingress.Check(); err != nil {
		return err
	}

	Opts = option
	Opts.NamespaceList = strings.Split(Opts.Namespaces, ",")

	return nil
}
