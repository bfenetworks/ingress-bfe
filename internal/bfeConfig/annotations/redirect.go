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

package annotations

import (
	"errors"
	"strconv"
)

const (
	redirectAnnotationPrefix          = BfeAnnotationPrefix + "redirect."
	defaultRedirectResponseStatusCode = 302
)

// the annotations related to how to set the location in the redirection response's header
const (
	RedirectURLSetAnnotation       = redirectAnnotationPrefix + "url-set"
	RedirectURLFromQueryAnnotation = redirectAnnotationPrefix + "url-from-query"
	RedirectURLPrefixAddAnnotation = redirectAnnotationPrefix + "url-prefix-add"
	RedirectSchemeSetSetAnnotation = redirectAnnotationPrefix + "scheme-set"
)

// RedirectResponseStatusAnnotation is used to set the status code of the redirection response manually
const RedirectResponseStatusAnnotation = redirectAnnotationPrefix + "response-status"

// GetRedirectAction try to parse the cmd and the param of the redirection action from the annotations
func GetRedirectAction(annotations map[string]string) (cmd, param string, err error) {
	switch {
	case annotations[RedirectURLSetAnnotation] != "":
		cmd, param = "URL_SET", annotations[RedirectURLSetAnnotation]

	case annotations[RedirectURLFromQueryAnnotation] != "":
		cmd, param = "URL_FROM_QUERY", annotations[RedirectURLFromQueryAnnotation]

	case annotations[RedirectURLPrefixAddAnnotation] != "":
		cmd, param = "URL_PREFIX_ADD", annotations[RedirectURLPrefixAddAnnotation]

	case annotations[RedirectSchemeSetSetAnnotation] != "":
		cmd, param = "SCHEME_SET", annotations[RedirectSchemeSetSetAnnotation]
	}
	return
}

func GetRedirectStatusCode(annotations map[string]string) (int, error) {
	statusCodeStr := annotations[RedirectResponseStatusAnnotation]
	if statusCodeStr == "" {
		return defaultRedirectResponseStatusCode, nil
	}

	statusCodeInt64, err := strconv.ParseInt(statusCodeStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if err != nil {
		return 0, errors.New("the annotation %s should be a integer")
	}
	return int(statusCodeInt64), nil
}
