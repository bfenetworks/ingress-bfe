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

import (
	"fmt"
	"strings"
)

const (
	CookieKey = "router.cookie"
	HeaderKey = "router.header"

	CookieAnnotation = BfeAnnotationPrefix + CookieKey
	HeaderAnnotation = BfeAnnotationPrefix + HeaderKey
)

func GetRouteExpression(annotations map[string]string) (string, error) {
	var primitive1 string
	var err error
	if primitive1, err = cookiePrimitive(annotations[CookieAnnotation]); err != nil {
		return "", err
	}

	var primitive2 string
	if primitive2, err = headerPrimitive(annotations[HeaderAnnotation]); err != nil {
		return "", err
	}

	if len(primitive1) > 0 && len(primitive2) > 0 {
		return primitive1 + "&&" + primitive2, nil
	}
	if len(primitive1) > 0 {
		return primitive1, nil
	}
	if len(primitive2) > 0 {
		return primitive2, nil
	}
	return "", nil
}

// cookiePrimitive generates bfe condition primitive for cookie match
func cookiePrimitive(cookie string) (string, error) {
	if len(cookie) == 0 {
		return "", nil
	}
	index := strings.Index(cookie, ":")
	if index == -1 || index == len(cookie)-1 {
		return "", fmt.Errorf("cookie annotation[%s] is illegal", cookie)
	}

	con := fmt.Sprintf("req_cookie_value_in(\"%s\", \"%v\", false)", strings.TrimSpace(cookie[:index]), strings.TrimSpace(cookie[index+1:]))
	return con, nil

}

// cookiePrimitive generates bfe condition primitive for header match
func headerPrimitive(header string) (string, error) {
	if len(header) == 0 {
		return "", nil
	}
	index := strings.Index(header, ":")
	if index == -1 || index == len(header)-1 {
		return "", fmt.Errorf("header annotation[%s] is illegal", header)
	}

	con := fmt.Sprintf("req_header_value_in(\"%s\", \"%v\", false)", strings.TrimSpace(header[:index]), strings.TrimSpace(header[index+1:]))
	return con, nil
}
