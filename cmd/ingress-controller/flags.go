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

	corev1 "k8s.io/api/core/v1"

	"github.com/bfenetworks/ingress-bfe/internal/option"
)

var (
	help           bool
	showVersion    bool
	namespaces     string
	ingressClass   string
	configPath     string
	reloadAddr     string
	metricsAddr    string
	probeAddr      string
	defaultBackend string
)

func initFlags() {
	flag.BoolVar(&help, "help", false, "Show help.")
	flag.BoolVar(&help, "h", false, "Show help.")

	flag.BoolVar(&showVersion, "version", false, "Show version of bfe-ingress-controller.")
	flag.BoolVar(&showVersion, "v", false, "Show version of bfe-ingress-controller.")

	flag.StringVar(&namespaces, "namespace", corev1.NamespaceAll, "Namespaces to watch, delimited by ','.")
	flag.StringVar(&namespaces, "n", corev1.NamespaceAll, "Namespaces to watch, delimited by ','.")

	flag.StringVar(&configPath, "bfe-config-path", option.ConfigPath, "Root directory of bfe configuration files.")
	flag.StringVar(&configPath, "c", option.ConfigPath, "Root directory of bfe configuration files.")

	flag.StringVar(&reloadAddr, "bfe-reload-address", option.ReloadAddr, "Address of bfe config reloading.")
	flag.StringVar(&ingressClass, "ingress-class", option.IngressClassName, "Class name of bfe ingress controller.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", option.MetricsBindAddress, "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", option.HealthProbeBindAddress, "The address the probe endpoint binds to.")
	flag.StringVar(&defaultBackend, "default-backend", option.DefaultBackend, "set default backend name, default backend is used if no any ingress rule matched, format namespace/name.")
}
