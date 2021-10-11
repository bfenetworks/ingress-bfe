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
package annotations

const (
	PriorityBasic        = 10
	PriorityHeader       = 20
	PriorityCookie       = 30
	PriorityCookieHeader = 40
)

func Priority(annotations map[string]string) int {
	_, ok1 := annotations[CookieAnnotation]
	_, ok2 := annotations[HeaderAnnotation]

	if ok1 && ok2 {
		return PriorityCookieHeader
	} else if ok1 {
		return PriorityCookie
	} else if ok2 {
		return PriorityHeader
	} else {
		return PriorityBasic
	}
}

func Equal(annotations1, annotations2 map[string]string) bool {
	if annotations1 == nil && annotations2 == nil {
		return true
	}

	return annotations1[CookieAnnotation] == annotations2[CookieAnnotation] &&
		annotations1[HeaderAnnotation] == annotations2[HeaderAnnotation]
}
