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

func BfePathType(pathType *networking.PathType) networking.PathType {
	if pathType == nil {
		return networking.PathTypePrefix
	}

	switch *pathType {
	// see: https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types
	case networking.PathTypeExact:
		return networking.PathTypeExact
	default:
		return networking.PathTypePrefix
	}
}
