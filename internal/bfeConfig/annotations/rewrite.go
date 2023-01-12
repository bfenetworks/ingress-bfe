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

type parseHandler func(string, string) ([]string, error)

// rawRewriteAction define struct of rewrite code and params.
// examples: {"HOST_SET": [{"params": "baidu.com", "when": "AfterLocation", "order": 1}]}.
// When BFE engine add new rewrite action callback points, user can set "when" field to other callback points
type rawRewriteAction map[string]rawRewriteActionParam
type rawRewriteActionParam []map[string]json.RawMessage

type RewriteActionParam struct {
	Params   []string
	Callback string
	Order    int
}

// RewriteAction contain rewrite action at different callback points.
type RewriteAction map[string]RewriteActionSegment

// RewriteActionSegment contains the rewrite actions at the same callback point.
type RewriteActionSegment map[string]*RewriteActionParam

// rewriteAnnotations contain the mapping of annotation to rewrite action.
var rewriteAnnotations = map[string]string{
	RewriteURLHostSetAnnotation:              "HOST_SET",
	RewriteURLHostFromPathAnnotation:         "HOST_SET_FROM_PATH_PREFIX",
	RewriteURLPathSetAnnotation:              "PATH_SET",
	RewriteURLPathPrefixAddAnnotation:        "PATH_PREFIX_ADD",
	RewriteURLPathPrefixTrimAnnotation:       "PATH_PREFIX_TRIM",
	RewriteURLPathPrefixStripAnnotation:      "PATH_STRIP",
	RewriteURLQueryAddAnnotation:             "QUERY_ADD",
	RewriteURLQueryDeleteAnnotation:          "QUERY_DEL",
	RewriteURLQueryRenameAnnotation:          "QUERY_RENAME",
	RewriteURLQueryDeleteAllExceptAnnotation: "QUERY_DEL_ALL_EXCEPT",
}

// parseParamHandlers contain the mapping of action to handler function.
var parseParamHandlers = map[string]parseHandler{
	"HOST_SET":                  hostSetHandler,
	"HOST_SET_FROM_PATH_PREFIX": hostSetFromPathPrefixHandler,
	"PATH_SET":                  pathHandler,
	"PATH_PREFIX_ADD":           pathHandler,
	"PATH_PREFIX_TRIM":          pathHandler,
	"PATH_STRIP":                pathPrefixStripHandler,
	"QUERY_ADD":                 queryAddAndRenameHandler,
	"QUERY_DEL":                 queryDeleteHandler,
	"QUERY_RENAME":              queryAddAndRenameHandler,
	"QUERY_DEL_ALL_EXCEPT":      queryDeleteAllExceptHandler,
}

var allowedCallbacks = map[string]interface{}{
	"AfterLocation": nil,
}

// GetRewriteAction parse the cmd and the param of the rewrite action from the annotations
func GetRewriteAction(annotations map[string]string) (RewriteAction, error) {
	// check annotation count constraint
	err := checkAnnotationCnt(annotations)
	if err != nil {
		return nil, err
	}
	// get raw actions
	rawAction, err := getRawRewriteAction(annotations)
	if err != nil {
		return nil, err
	}
	if len(rawAction) == 0 {
		return nil, nil
	}
	// parse actions and divide action by callback point
	return parseRawRewriteAction(rawAction)
}

// GetPathStripPrefix get the particular path prefix from strip segment length.
// examples: for path "/bar/test", if the strip length is 1, this strip prefix is /bar, but if the strip length is 3, this strip prefix is an empty string.
func GetPathStripPrefix(path string, length string) (string, error) {
	l, err := checkStripLength(length)
	if err != nil {
		return "", err
	}
	// remove wildcard suffix
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

// CheckAllowedCallBack check callback point, only support AfterLocation in this version.
func CheckAllowedCallBack(cb string) error {
	_, ok := allowedCallbacks[cb]
	if !ok {
		return fmt.Errorf("not allowed callback: %s", cb)
	}
	return nil
}

func getRawRewriteAction(annotations map[string]string) (rawRewriteAction, error) {
	actions := make(rawRewriteAction)
	for annot, v := range annotations {
		if cmd, ok := rewriteAnnotations[annot]; ok {
			param := make(rawRewriteActionParam, 0)
			err := json.Unmarshal([]byte(v), &param)
			if err != nil {
				return nil, fmt.Errorf("annotation %s's param is illegal, error: %s", annot, err)
			}
			actions[cmd] = param
		}
	}
	return actions, nil
}

func parseRawRewriteAction(rawAction rawRewriteAction) (RewriteAction, error) {
	// actions: callback -> action -> actionParam
	actions := make(map[string]RewriteActionSegment, len(rawAction))
	for cmd, rewriteParamList := range rawAction {
		for _, rewriteParam := range rewriteParamList {
			param, e := getRewriteActionParam(cmd, rewriteParam)
			if e != nil {
				return nil, e
			}
			if _, ok := actions[param.Callback]; !ok {
				actions[param.Callback] = make(map[string]*RewriteActionParam)
			}
			if _, ok := actions[param.Callback][cmd]; ok {
				return nil, errors.New("setting a rewrite-action with duplicate Callback points is not allowed")
			}
			actions[param.Callback][cmd] = param
		}
	}
	return actions, nil
}

// getRewriteActionParam parse param, order and callback property from annotation.
func getRewriteActionParam(cmd string, actionParam map[string]json.RawMessage) (*RewriteActionParam, error) {
	rawParam := parseRawJsonToString(actionParam[paramKey])
	if rawParam == "" {
		return nil, errors.New("missing \"params\" field in rewrite-url action")
	}

	if _, ok := parseParamHandlers[cmd]; !ok {
		return nil, fmt.Errorf("unsupported annotation for rewrite action: %s", cmd)
	}
	handler := parseParamHandlers[cmd]
	params, err := handler(cmd, rawParam)
	if err != nil {
		return nil, err
	}

	rawOrder := parseRawJsonToString(actionParam[orderKey])
	var order int64
	if rawOrder != "" {
		order, err = strconv.ParseInt(rawOrder, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	callback := parseRawJsonToString(actionParam[callBackKey])
	if callback != "" {
		err = CheckAllowedCallBack(callback)
		if err != nil {
			return nil, err
		}
	} else {
		callback = DefaultCallBackPoint
	}

	return &RewriteActionParam{
		params, callback, int(order),
	}, nil
}

func hostSetHandler(_, param string) ([]string, error) {
	if err := checkHost(param); err != nil {
		return nil, fmt.Errorf("invalid host name, error: %s", err)
	}
	return []string{param}, nil
}

func hostSetFromPathPrefixHandler(_, param string) ([]string, error) {
	if param != "true" {
		return nil, fmt.Errorf("invalid host-set-from-path param: %s, which should be true", param)
	}
	return []string{}, nil
}

func pathHandler(cmd, param string) ([]string, error) {
	if err := checkPath(param); err != nil {
		return nil, err
	}
	if cmd == "PATH_PREFIX_ADD" && !strings.HasSuffix(param, "/") {
		param += "/"
	}
	return []string{param}, nil
}

func pathPrefixStripHandler(_, param string) ([]string, error) {
	_, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid path-frefix-strip param: %s, error: %s", param, err)
	}
	return []string{param}, nil
}

func queryAddAndRenameHandler(cmd, param string) ([]string, error) {
	paramMaps := make(map[string]string)
	err := json.Unmarshal([]byte(param), &paramMaps)
	if err != nil {
		return nil, err
	}
	if len(paramMaps) == 0 {
		return nil, fmt.Errorf("the param of %s can not be empty", cmd)
	}
	paramList := make([]string, 0)
	for k, v := range paramMaps {
		paramList = append(paramList, k, v)
	}
	return paramList, nil
}

func queryDeleteHandler(cmd, param string) ([]string, error) {
	paramList := make([]string, 0)
	err := json.Unmarshal([]byte(param), &paramList)
	if err != nil {
		return nil, err
	}
	if len(paramList) == 0 {
		return nil, fmt.Errorf("the param of %s can not be empty", cmd)
	}
	return paramList, nil
}

func queryDeleteAllExceptHandler(_, param string) ([]string, error) {
	return []string{param}, nil
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

func checkStripLength(rawLength string) (int64, error) {
	l, err := strconv.ParseInt(rawLength, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("parse path strip length error: %s", err)
	}
	if l <= 0 {
		return -1, fmt.Errorf("invalid path strip length: %v, which should greater than 0", l)
	}
	return l, nil
}
