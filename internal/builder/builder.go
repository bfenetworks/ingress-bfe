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

type BfeConfigDumper interface {
	Dump() error
}

type BfeConfigCache interface {
	Submit(ingress *networking.Ingress) error
	Rollback(ingress *networking.Ingress) error
}

type BfeConfigReloader interface {
	Reload() error
}

type BfeConfigBuilder interface {
	BfeConfigReloader
	BfeConfigCache
	BfeConfigDumper
	Build() error
}
