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
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

const (
	ConfigPath      = "/bfe/conf/"
	ReloadAddr      = "localhost:8421"
	reloadInterval  = 3 * time.Second
	reloadUrlPrefix = "http://%s/reload/"

	FilePerm os.FileMode = 0744

	MetricsBindAddress     = ":9080"
	HealthProbeBindAddress = ":9081"

	// used in ingress annotation as value of key kubernetes.io/ingress.class
	IngressClassName = "bfe"

	// used in IngressClass resource as value of controller
	ControllerName = "bfe-networks.com/ingress-controller"

	// default backend
	DefaultBackend = ""
)

type Options struct {
	Namespaces      []string
	IngressClass    string
	ControllerName  string
	ReloadUrl       string
	ConfigPath      string
	MetricsAddr     string
	HealthProbeAddr string
	ReloadInterval  time.Duration
	DefaultBackend  string
}

var (
	Opts *Options
)

func SetOptions(namespaces, class, configPath, reloadAddr, metricsAddr, probeAddr, defaultBackend string) error {
	if len(defaultBackend) > 0 {
		names := strings.Split(defaultBackend, string(types.Separator))
		if len(names) != 2 {
			return fmt.Errorf("invalid command line argument default-backend: %s", defaultBackend)
		}
	}

	if !strings.HasSuffix(configPath, "/") {
		configPath = configPath + "/"
	}

	Opts = &Options{
		Namespaces:      strings.Split(namespaces, ","),
		IngressClass:    class,
		ControllerName:  ControllerName,
		ReloadUrl:       fmt.Sprintf(reloadUrlPrefix, reloadAddr),
		ConfigPath:      configPath,
		MetricsAddr:     metricsAddr,
		HealthProbeAddr: probeAddr,
		ReloadInterval:  reloadInterval,
		DefaultBackend:  defaultBackend,
	}

	return nil
}
