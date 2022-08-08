// Copyright (c) 2022 The BFE Authors.
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

package modules

import (
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/modules/redirect"
	netv1 "k8s.io/api/networking/v1"
)

// BFEModuleConfig is an abstraction of the BFE module configuration.
// The ConfigBuilder will call the corresponding function in this interface when update/delete ingresses or reload the BFE Engine.
type BFEModuleConfig interface {
	// UpdateIngress uses the ingress to update the BFEModuleConfig
	UpdateIngress(ingress *netv1.Ingress) error

	// DeleteIngress delete everything related to the ingress from the BFEModuleConfig
	DeleteIngress(ingressNamespace, ingressName string)

	// Reload dumps the data in the BFEModuleConfig to the corresponding BFE conf files on the disk
	Reload() error

	// Name returns the name of the BFEModuleConfig
	Name() string
}

func InitBFEModules(version string) []BFEModuleConfig {
	var modules []BFEModuleConfig
	// mod_redirect
	modules = append(modules, redirect.NewRedirectConfig(version))
	return modules
}
