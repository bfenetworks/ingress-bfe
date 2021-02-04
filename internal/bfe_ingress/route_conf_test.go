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

package bfe_ingress

import (
	"reflect"
	"testing"
)

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
)

func Test_buildCondition(t *testing.T) {
	type args struct {
		host    string
		path    string
		exConds []BfeAnnotation
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 int
	}{
		// TODO: Add test cases.
		{
			name: "host + path",
			args: args{
				host:    "example.com",
				path:    "/foo",
				exConds: nil,
			},
			want:  "req_host_in(\"example.com\") && req_path_element_prefix_in(\"/foo\", false)",
			want1: ConditionTypeContainHostPrefixPath,
		},
		{
			name: "only host",
			args: args{
				host:    "example.com",
				path:    "",
				exConds: nil,
			},
			want:  "req_host_in(\"example.com\")",
			want1: ConditionTypeContainOnlyHost,
		},
		{
			name: "only path",
			args: args{
				host:    "",
				path:    "/foo",
				exConds: nil,
			},
			want:  "req_path_element_prefix_in(\"/foo\", false)",
			want1: ConditionTypeContainOnlyPrefixPath,
		},
		{
			name: "no host and path",
			args: args{
				host:    "",
				path:    "",
				exConds: nil,
			},
			want:  "default_t()",
			want1: ConditionTypeContainNoHostPath,
		},
		{
			name: "host + path + header",
			args: args{
				host: "example.com",
				path: "/foo",
				exConds: []BfeAnnotation{
					&headerAnnotation{annotationStr: "HeaderTest: test"},
				},
			},
			want:  "req_host_in(\"example.com\") && req_path_element_prefix_in(\"/foo\", false) && req_header_value_in(\"HeaderTest\", \"test\", false)",
			want1: ConditionTypeContainHostPrefixPath,
		},
		{
			name: "host + path + cookie",
			args: args{
				host: "example.com",
				path: "/foo",
				exConds: []BfeAnnotation{
					&cookieAnnotation{annotationStr: "CookieTest: test"},
				},
			},
			want:  "req_host_in(\"example.com\") && req_path_element_prefix_in(\"/foo\", false) && req_cookie_value_in(\"CookieTest\", \"test\", false)",
			want1: ConditionTypeContainHostPrefixPath,
		},
		{
			name: "no host and path,  with cookie",
			args: args{
				host: "",
				path: "",
				exConds: []BfeAnnotation{
					&cookieAnnotation{annotationStr: "CookieTest: test"},
				},
			},
			want:  "default_t()",
			want1: ConditionTypeContainNoHostPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildCondition(tt.args.host, tt.args.path, nil, tt.args.exConds)
			if got != tt.want {
				t.Errorf("buildCondition() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("buildCondition() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_sortRules(t *testing.T) {
	var cluster = "test"
	var condHostPathHeaderCookie = "req_host_in(\"example.com\") && req_path_element_prefix_in(\"/foo\", false) && req_cookie_value_in(\"CookieTest\", \"test\", false) && req_header_value_in(\"HeaderTest\", \"test\", false)"
	var condHostPathCookie = "req_host_in(\"example.com\") && req_path_element_prefix_in(\"/foo\", false) && req_cookie_value_in(\"CookieTest\", \"test\", false)"
	var condHostPathHeader = "req_host_in(\"example.com\") && req_path_element_prefix_in(\"/foo\", false) && req_header_value_in(\"HeaderTest\", \"test\", false)"
	var condHostPath = "req_host_in(\"example.com\") && req_path_element_prefix_in(\"/foo\", false)"
	var condHost = "req_host_in(\"example.com\")"
	var condPath = "req_path_element_prefix_in(\"/foo\", false)"
	var condHostCookie = "req_host_in(\"example.com\") && req_cookie_value_in(\"CookieTest\", \"test\", false)"
	var condPathCookie = "req_host_in(\"example.com\") && req_cookie_value_in(\"CookieTest\", \"test\", false)"
	var defaultStr = "default_t()"

	var ruleDefault = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &defaultStr,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host: "",
			Path: "",
			Annotations: []BfeAnnotation{
				&cookieAnnotation{annotationStr: "CookieTest: test"},
				&headerAnnotation{annotationStr: "HeaderTest: test"},
			},
		},
		ConditionType: ConditionTypeContainHostPrefixPath,
	}

	var ruleHostPathHeaderCookie = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condHostPathHeaderCookie,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host: "example.com",
			Path: "/foo",
			Annotations: []BfeAnnotation{
				&cookieAnnotation{annotationStr: "CookieTest: test"},
				&headerAnnotation{annotationStr: "HeaderTest: test"},
			},
		},
		ConditionType: ConditionTypeContainHostPrefixPath,
	}

	var ruleHostPathCookie = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condHostPathCookie,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host: "example.com",
			Path: "/foo",
			Annotations: []BfeAnnotation{
				&cookieAnnotation{annotationStr: "CookieTest: test"},
			},
		},
		ConditionType: ConditionTypeContainHostPrefixPath,
	}

	var ruleHostPathHeader = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condHostPathHeader,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host: "example.com",
			Path: "/foo",
			Annotations: []BfeAnnotation{
				&headerAnnotation{annotationStr: "HeaderTest: test"},
			},
		},
		ConditionType: ConditionTypeContainHostPrefixPath,
	}

	var ruleHostPath = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condHostPath,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host:        "example.com",
			Path:        "/foo",
			Annotations: nil,
		},
		ConditionType: ConditionTypeContainHostPrefixPath,
	}

	var ruleHost = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condHost,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host:        "example.com",
			Path:        "",
			Annotations: nil,
		},
		ConditionType: ConditionTypeContainOnlyHost,
	}

	var rulePath = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condPath,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host:        "",
			Path:        "/foo",
			Annotations: nil,
		},
		ConditionType: ConditionTypeContainOnlyPrefixPath,
	}

	var ruleHostCookie = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condHostCookie,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host: "example.com",
			Path: "",
			Annotations: []BfeAnnotation{
				&cookieAnnotation{annotationStr: "CookieTest: test"},
			},
		},
		ConditionType: ConditionTypeContainOnlyHost,
	}
	var rulePathCookie = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.RouteRuleFile{
			Cond:        &condPathCookie,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host: "",
			Path: "/foo",
			Annotations: []BfeAnnotation{
				&cookieAnnotation{annotationStr: "CookieTest: test"},
			},
		},
		ConditionType: ConditionTypeContainOnlyPrefixPath,
	}

	type args struct {
		routeRuleFiles []ingressRouteRuleFile
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				routeRuleFiles: []ingressRouteRuleFile{
					ruleDefault,
					ruleHost,
					ruleHostCookie,
					rulePath,
					rulePathCookie,
					ruleHostPath,
					ruleHostPathCookie,
					ruleHostPathHeader,
					ruleHostPathHeaderCookie,
				},
			},
			want: args{
				routeRuleFiles: []ingressRouteRuleFile{
					ruleHostPathHeaderCookie,
					ruleHostPathCookie,
					ruleHostPathHeader,
					ruleHostPath,
					ruleHostCookie,
					ruleHost,
					rulePathCookie,
					rulePath,
					ruleDefault,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortRules(tt.args.routeRuleFiles)
			if reflect.DeepEqual(tt.args.routeRuleFiles, tt.want.routeRuleFiles) {
				t.Errorf("buildCondition() got1 = %v, want %v", tt.args.routeRuleFiles, tt.want.routeRuleFiles)
			}
		})
	}
}
