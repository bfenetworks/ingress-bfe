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

//Package annotations defines the bfe-ingress annotation's format and converter.
//This file is related to rewrite-url annotations.
package annotations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const rewriteAnnotationPrefix = BfeAnnotationPrefix + "rewrite-url."

// The annotation related to rewrite host action.
const (
	RewriteURLHostSetAnnotation      = rewriteAnnotationPrefix + "host"
	RewriteURLHostFromPathAnnotation = rewriteAnnotationPrefix + "host-from-path-prefix"
)

// The annotation related to rewrite path action.
const (
	RewriteURLPathSetAnnotation         = rewriteAnnotationPrefix + "path"
	RewriteURLPathPrefixAddAnnotation   = rewriteAnnotationPrefix + "path-prefix-add"
	RewriteURLPathPrefixTrimAnnotation  = rewriteAnnotationPrefix + "path-prefix-trim"
	RewriteURLPathPrefixStripAnnotation = rewriteAnnotationPrefix + "path-prefix-strip"
)

// The annotation related to rewrite query action.
const (
	RewriteURLQueryAddAnnotation             = rewriteAnnotationPrefix + "query-add"
	RewriteURLQueryDeleteAnnotation          = rewriteAnnotationPrefix + "query-delete"
	RewriteURLQueryRenameAnnotation          = rewriteAnnotationPrefix + "query-rename"
	RewriteURLQueryDeleteAllExceptAnnotation = rewriteAnnotationPrefix + "query-delete-all-except"
)

const (
	callBackKey = "when"
	paramKey    = "params"
	orderKey    = "order"
)

const DefaultCallBackPoint = "AfterLocation"

// transRewriteAnnotationKeyToAction convert bfe-ingress rewrite annotation key to BFE engine rewrite action key
func transRewriteAnnotationKeyToAction(annot string) (ac string) {
	switch annot {
	case RewriteURLHostSetAnnotation:
		ac = "HOST_SET"
	case RewriteURLHostFromPathAnnotation:
		ac = "HOST_SET_FROM_PATH_PREFIX"
	case RewriteURLPathSetAnnotation:
		ac = "PATH_SET"
	case RewriteURLPathPrefixAddAnnotation:
		ac = "PATH_PREFIX_ADD"
	case RewriteURLPathPrefixTrimAnnotation:
		ac = "PATH_PREFIX_TRIM"
	case RewriteURLPathPrefixStripAnnotation:
		ac = "PATH_STRIP"
	case RewriteURLQueryAddAnnotation:
		ac = "QUERY_ADD"
	case RewriteURLQueryDeleteAnnotation:
		ac = "QUERY_DEL"
	case RewriteURLQueryRenameAnnotation:
		ac = "QUERY_RENAME"
	case RewriteURLQueryDeleteAllExceptAnnotation:
		ac = "QUERY_DEL_ALL_EXCEPT"
	default:
		ac = ""
	}
	return
}

// RewriteAction define struct of rewrite annotation and Params.
// examples: {"HOST_SET": [{"params": "baidu.com", "when": "AfterLocation", "order": "1"}]}.
// When BFE engine add new rewrite action Callback points, user can set "when" field to other Callback points
type RewriteAction map[string]RewriteActionParamList
type RewriteActionParamList []map[string]json.RawMessage

type ActionParam struct {
	Params   []string
	Callback string
	Order    int
}

// parseActionParam convert param to raw param to param slice.
func parseActionParam(cmd, param string) ([]string, error) {
	switch cmd {
	case "HOST_SET":
		if err := checkHost(param); err != nil {
			return nil, fmt.Errorf("invalid host name, error: %s", err)
		}
		return []string{param}, nil
	case "HOST_SET_FROM_PATH_PREFIX":
		if param == "" || strings.ToLower(param) != "true" && strings.ToLower(param) != "t" {
			return nil, fmt.Errorf("invalid host-set-from-path param: %s, which should be true or t", param)
		}
		return []string{}, nil
	case "PATH_PREFIX_ADD":
		if err := checkPath(param); err != nil {
			return nil, err
		}
		if !strings.HasSuffix(param, "/") {
			param += "/"
		}
		return []string{param}, nil
	case "PATH_SET", "PATH_PREFIX_TRIM":
		if err := checkPath(param); err != nil {
			return nil, err
		}
		return []string{param}, nil
	case "PATH_STRIP":
		_, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid path-frefix-strip param: %s, error: %s", param, err)
		}
		return []string{param}, nil
	case "QUERY_ADD", "QUERY_RENAME":
		paramMaps := make(map[string]string)
		err := json.Unmarshal([]byte(param), &paramMaps)
		if err != nil {
			return nil, err
		}
		if len(paramMaps) == 0 {
			return nil, fmt.Errorf("the param of annotation %s can not be empty", cmd)
		}
		paramList := make([]string, 0)
		for k, v := range paramMaps {
			paramList = append(paramList, k, v)
		}
		return paramList, nil
	case "QUERY_DEL":
		paramList := make([]string, 0)
		err := json.Unmarshal([]byte(param), &paramList)
		if err != nil {
			return nil, err
		}
		if len(paramList) == 0 {
			return nil, fmt.Errorf("the param of annotation %s can not be empty", cmd)
		}
		return paramList, nil
	case "QUERY_DEL_ALL_EXCEPT":
		return []string{param}, nil
	default:
		return nil, fmt.Errorf("unsupported annotation for rewrite action: %s", cmd)
	}
}

// parseRawJsonToString convert raw message to string and remove the beginning and ending double quote(") in the string.
func parseRawJsonToString(raw json.RawMessage) string {
	param := string(raw)

	if strings.HasPrefix(param, "\"") {
		param = strings.TrimPrefix(param, "\"")
	}

	if strings.HasSuffix(param, "\"") {
		param = strings.TrimSuffix(param, "\"")
	}

	return param
}

// getActionParam parse "param","order","callback" property from annotation.
func getActionParam(cmd string, actionParam map[string]json.RawMessage) (*ActionParam, error) {
	param := parseRawJsonToString(actionParam[paramKey])
	if param == "" {
		return nil, errors.New("missing \"Params\" field in rewrite-url action")
	}

	params, err := parseActionParam(cmd, param)
	if err != nil {
		return nil, err
	}

	orderStr := parseRawJsonToString(actionParam[orderKey])
	var order int64
	if orderStr != "" {
		order, err = strconv.ParseInt(orderStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	callback := parseRawJsonToString(actionParam[callBackKey])
	if callback != "" {
		err = CheckCallBackPoint(callback)
		if err != nil {
			return nil, err
		}
	} else {
		callback = DefaultCallBackPoint
	}
	return &ActionParam{
		params, callback, int(order),
	}, nil
}

// GetRewriteActions try to parse the cmd and the param of the rewrite action from the annotations
func GetRewriteActions(annotations map[string]string) (map[string]map[string]*ActionParam, error) {
	err := checkAnnotationCnt(annotations)
	if err != nil {
		return nil, err
	}

	rawActions := make(RewriteAction)
	for annot, v := range annotations {
		if cmd := transRewriteAnnotationKeyToAction(annot); cmd != "" {
			param := make(RewriteActionParamList, 0)
			err := json.Unmarshal([]byte(v), &param)
			if err != nil {
				return nil, fmt.Errorf("annotation %s's param is illegal, error: %s", annot, err)
			}
			rawActions[cmd] = param
		}
	}

	// actions: Callback -> action -> actionParam
	actions := make(map[string]map[string]*ActionParam, len(rawActions))
	for cmd, rewriteParamList := range rawActions {
		for _, rewriteParam := range rewriteParamList {
			param, e := getActionParam(cmd, rewriteParam)
			if e != nil {
				return nil, e
			}

			if _, ok := actions[param.Callback]; !ok {
				actions[param.Callback] = make(map[string]*ActionParam)
			}

			if _, ok := actions[param.Callback][cmd]; ok {
				return nil, errors.New("setting a rewrite-action with duplicate Callback points is not allowed")
			}

			actions[param.Callback][cmd] = param
		}
	}
	return actions, nil
}

// ConvertStripParam get the particular path prefix from strip segment length.
// examples: for path "/bar/test", if the strip length is 1, this strip prefix is /bar, but if the strip length is 3, this strip prefix is an empty string.
func ConvertStripParam(param string, path string) (string, error) {
	l, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return "", err
	}
	if l <= 0 {
		return "", fmt.Errorf("invalid path strip length:%v", l)
	}
	// remove suffix
	if strings.Count(path, "*") == 1 && strings.HasSuffix(path, "*") {
		path = path[:len(path)-1]
	}
	segs := strings.Split(path, "/")
	var prefix string
	if len(segs) > int(l) {
		prefix = strings.Join(segs[:l+1], "/")
	}
	return prefix, nil
}

func getRelatedAnnotationCnt(annots map[string]string, keys []string) int {
	cnt := 0
	for _, key := range keys {
		if _, ok := annots[key]; ok {
			cnt++
		}
	}
	return cnt
}

// checkAnnotationCnt check the annotation's count constraint.
func checkAnnotationCnt(annots map[string]string) error {
	// check host annotation cnt
	hostAnnotations := []string{
		RewriteURLHostSetAnnotation,
		RewriteURLHostFromPathAnnotation,
	}
	if getRelatedAnnotationCnt(annots, hostAnnotations) == 2 {
		return fmt.Errorf("setting annotations %s and %s at the same time is not allowed", RewriteURLHostSetAnnotation, RewriteURLHostFromPathAnnotation)
	}

	// check path annotation cnt
	pathAnnotations := []string{
		RewriteURLPathSetAnnotation,
		RewriteURLPathPrefixAddAnnotation,
		RewriteURLPathPrefixTrimAnnotation,
	}
	if annots[RewriteURLPathSetAnnotation] != "" && getRelatedAnnotationCnt(annots, pathAnnotations) > 1 {
		return errors.New("when set a fixed url-path annotation, setting path-prefix-add or path-prefix-trim annotation is not allowed")
	}
	return nil
}

func checkHost(host string) error {
	// wildcard hostname: started with "*." is allowed
	if strings.Count(host, "*") > 1 || (strings.Count(host, "*") == 1 && !strings.HasPrefix(host, "*.")) {
		return fmt.Errorf("wildcard host[%s] is illegal, should start with *. ", host)
	}
	return nil
}

func checkPath(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("path is not set")
	}

	if strings.ContainsAny(path, "*") {
		return fmt.Errorf("path[%s] is illegal", path)
	}

	if _, err := url.Parse(path); err != nil {
		return err
	}
	return nil
}

// CheckCallBackPoint check callback point, only support AfterLocation in this version.
func CheckCallBackPoint(cb string) error {
	switch cb {
	case DefaultCallBackPoint:
		return nil
	default:
		return fmt.Errorf("%s callback point in rewrite-url action is not supported", cb)
	}
}
