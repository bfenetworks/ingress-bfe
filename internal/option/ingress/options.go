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

package ingress

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

const (
	enableIngress = true

	configPath      = "/bfe/conf/"
	bfeBinary       = "/bfe/bin/bfe"
	reloadAddr      = "localhost:8421"
	reloadInterval  = 3 * time.Second
	reloadUrlPrefix = "http://%s/reload/"

	filePerm os.FileMode = 0744

	// used in ingress annotation as value of key kubernetes.io/ingress.class
	ingressClassName = "bfe"

	// used in IngressClass resource as value of controller
	controllerName = "bfe-networks.com/ingress-controller"

	// default backend
	defaultBackend = ""
)

type Options struct {
	EnableIngress  bool
	IngressClass   string
	ControllerName string
	ReloadAddr     string
	ReloadUrl      string
	BfeBinary      string
	ConfigPath     string
	FilePerm       os.FileMode
	ReloadInterval time.Duration
	DefaultBackend string
}

func NewOptions() *Options {
	return &Options{
		EnableIngress:  enableIngress,
		IngressClass:   ingressClassName,
		ControllerName: controllerName,
		ReloadAddr:     reloadAddr,
		BfeBinary:      bfeBinary,
		ConfigPath:     configPath,
		FilePerm:       filePerm,
		ReloadInterval: reloadInterval,
		DefaultBackend: defaultBackend,
	}
}

func (opts *Options) Check() error {
	if !opts.EnableIngress {
		return nil
	}

	if len(opts.DefaultBackend) > 0 {
		names := strings.Split(opts.DefaultBackend, string(types.Separator))
		if len(names) != 2 {
			return fmt.Errorf("invalid command line argument default-backend: %s", opts.DefaultBackend)
		}
	}
	if len(opts.BfeBinary) > 0 {
		opts.ConfigPath = filepath.Dir(filepath.Dir(opts.BfeBinary)) + "/conf"
	}

	if !strings.HasSuffix(opts.ConfigPath, "/") {
		opts.ConfigPath = opts.ConfigPath + "/"
	}

	opts.ReloadUrl = fmt.Sprintf(reloadUrlPrefix, opts.ReloadAddr)
	return nil
}
