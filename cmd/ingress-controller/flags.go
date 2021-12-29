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

package main

import (
	"flag"

	"github.com/bfenetworks/ingress-bfe/internal/option"
)

var (
	help        bool
	showVersion bool

	opts *option.Options = option.NewOptions()
)

func initFlags() {
	flag.BoolVar(&help, "help", false, "Show help.")
	flag.BoolVar(&help, "h", false, "Show help.")

	flag.BoolVar(&showVersion, "version", false, "Show version of bfe-ingress-controller.")
	flag.BoolVar(&showVersion, "v", false, "Show version of bfe-ingress-controller.")

	flag.StringVar(&opts.Namespaces, "namespace", opts.Namespaces, "Namespaces to watch, delimited by ','.")
	flag.StringVar(&opts.Namespaces, "n", opts.Namespaces, "Namespaces to watch, delimited by ','.")

	flag.StringVar(&opts.MetricsAddr, "metrics-bind-address", opts.MetricsAddr, "The address the metric endpoint binds to.")
	flag.StringVar(&opts.HealthProbeAddr, "health-probe-bind-address", opts.HealthProbeAddr, "The address the probe endpoint binds to.")
	flag.StringVar(&opts.ClusterName, "k8s-cluster-name", opts.ClusterName, "k8s cluster name")

	flag.StringVar(&opts.Ingress.ConfigPath, "bfe-config-path", opts.Ingress.ConfigPath, "Root directory of bfe configuration files.")
	flag.StringVar(&opts.Ingress.ConfigPath, "c", opts.Ingress.ConfigPath, "Root directory of bfe configuration files.")
	flag.StringVar(&opts.Ingress.BfeBinary, "bfe-binary", opts.Ingress.BfeBinary, "Absolute path of BFE binary. If set, <bfe-config-path> is overwritten by <bfe-binary>/../conf")
	flag.StringVar(&opts.Ingress.BfeBinary, "b", opts.Ingress.BfeBinary, "Absolute path of BFE binary. If set, <bfe-config-path> is overwritten by <bfe-binary>/../conf,")
	flag.StringVar(&opts.Ingress.ReloadAddr, "bfe-reload-address", opts.Ingress.ReloadAddr, "Address of bfe config reloading.")
	flag.StringVar(&opts.Ingress.IngressClass, "ingress-class", opts.Ingress.IngressClass, "Class name of bfe ingress controller.")
	flag.StringVar(&opts.Ingress.DefaultBackend, "default-backend", opts.Ingress.DefaultBackend, "set default backend name, default backend is used if no any ingress rule matched, format namespace/name.")

}
