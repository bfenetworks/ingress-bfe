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
	networking "k8s.io/api/networking/v1beta1"
)

// BfeConfigCache is interface for cacheable BfeConfigBuilder,
// which can cache ingresses one-by-one for specific BFE config
type BfeConfigCache interface {
	/*
		Submit an ingress resource to cache of specific BFE config.

		All ingress resources will be submitted in sequence.
		It supposed that: Submit with error won't change BfeConfigCache.
	*/
	Submit(ingress *networking.Ingress) error

	/*
		Rollback is reverse operation of Submit.
		It changes cache of BFE config with a submitted ingress resource, as if it hadn't been submitted.
	*/
	Rollback(ingress *networking.Ingress) error
}

// BfeConfigDumper dumps specific BFE config
type BfeConfigDumper interface {
	Dump() error
}

// BfeConfigReloader reloads specific BFE config
// usually work with BfeConfigDumper( reload after dump)
type BfeConfigReloader interface {
	Reload() error
}

// BfeConfigBuilder build specific BFE config
type BfeConfigBuilder interface {
	// cache information from ingresses
	BfeConfigCache

	// Build builds BFE config, usually use information cached before
	Build() error

	// dump BFE config for subsequent use (e.g. reloading, troubleshooting ...)
	BfeConfigDumper

	// reload BFE config
	BfeConfigReloader
}
