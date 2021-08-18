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
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/stretchr/testify/assert"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	DemoHostExact    = "example.host.com"
	DemoHostWildcard = "*.example.host.com"
)

func Test_buildCondition(t *testing.T) {
	type args struct {
		host    string
		path    string
		exConds []BfeAnnotation
	}
	tests := []struct {
		name     string
		args     args
		want     string
		wantType int
	}{
		{
			name: "host + path",
			args: args{
				host:    "example.com",
				path:    "/foo",
				exConds: nil,
			},
			want:     `req_host_in("example.com") && (req_path_in("/foo", false) || req_path_prefix_in("/foo/", false))`,
			wantType: ConditionTypeContainExactHostPrefixPath,
		},
		{
			name: "only host",
			args: args{
				host:    "example.com",
				path:    "",
				exConds: nil,
			},
			want:     "req_host_in(\"example.com\")",
			wantType: ConditionTypeContainOnlyExactHost,
		},
		{
			name: "only path",
			args: args{
				host:    "",
				path:    "/foo",
				exConds: nil,
			},
			want:     `(req_path_in("/foo", false) || req_path_prefix_in("/foo/", false))`,
			wantType: ConditionTypeContainOnlyPrefixPath,
		},
		{
			name: "path end with /",
			args: args{
				host:    "",
				path:    "/foo/",
				exConds: nil,
			},
			want:     `(req_path_in("/foo", false) || req_path_prefix_in("/foo/", false))`,
			wantType: ConditionTypeContainOnlyPrefixPath,
		},
		{
			name: "no host and path",
			args: args{
				host:    "",
				path:    "",
				exConds: nil,
			},
			want:     "default_t()",
			wantType: ConditionTypeContainNoHostPath,
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
			want:     `req_host_in("example.com") && (req_path_in("/foo", false) || req_path_prefix_in("/foo/", false)) && req_header_value_in("HeaderTest", "test", false)`,
			wantType: ConditionTypeContainExactHostPrefixPath,
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
			want:     `req_host_in("example.com") && (req_path_in("/foo", false) || req_path_prefix_in("/foo/", false)) && req_cookie_value_in("CookieTest", "test", false)`,
			wantType: ConditionTypeContainExactHostPrefixPath,
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
			want:     "default_t()",
			wantType: ConditionTypeContainNoHostPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildCondition(tt.args.host, tt.args.path, nil, tt.args.exConds)
			if got != tt.want {
				t.Errorf("buildCondition() got = %v \n want %v", got, tt.want)
			}
			if got1 != tt.wantType {
				t.Errorf("buildCondition() got1 = %v \n want %v", got1, tt.wantType)
			}
		})
	}
}

func Test_sortRules(t *testing.T) {
	var cluster = "test"
	var condHostPathHeaderCookie = "req_host_in(\"example.com\") && req_path_prefix_in(\"/foo\", false) && req_cookie_value_in(\"CookieTest\", \"test\", false) && req_header_value_in(\"HeaderTest\", \"test\", false)"
	var condHostPathCookie = "req_host_in(\"example.com\") && req_path_prefix_in(\"/foo\", false) && req_cookie_value_in(\"CookieTest\", \"test\", false)"
	var condHostPathHeader = "req_host_in(\"example.com\") && req_path_prefix_in(\"/foo\", false) && req_header_value_in(\"HeaderTest\", \"test\", false)"
	var condHostPath = "req_host_in(\"example.com\") && req_path_prefix_in(\"/foo\", false)"
	var condHost = "req_host_in(\"example.com\")"
	var condPath = "req_path_prefix_in(\"/foo\", false)"
	var condHostCookie = "req_host_in(\"example.com\") && req_cookie_value_in(\"CookieTest\", \"test\", false)"
	var condPathCookie = "req_host_in(\"example.com\") && req_cookie_value_in(\"CookieTest\", \"test\", false)"
	var defaultStr = "default_t()"

	var ruleDefault = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
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
		ConditionType: ConditionTypeContainExactHostPrefixPath,
	}

	var ruleHostPathHeaderCookie = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
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
		ConditionType: ConditionTypeContainExactHostPrefixPath,
	}

	var ruleHostPathCookie = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
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
		ConditionType: ConditionTypeContainExactHostPrefixPath,
	}

	var ruleHostPathHeader = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
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
		ConditionType: ConditionTypeContainExactHostPrefixPath,
	}

	var ruleHostPath = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
			Cond:        &condHostPath,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host:        "example.com",
			Path:        "/foo",
			Annotations: nil,
		},
		ConditionType: ConditionTypeContainExactHostPrefixPath,
	}

	var ruleHost = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
			Cond:        &condHost,
			ClusterName: &cluster,
		},
		RawRuleInfo: ingressRawRuleInfo{
			Host:        "example.com",
			Path:        "",
			Annotations: nil,
		},
		ConditionType: ConditionTypeContainOnlyExactHost,
	}

	var rulePath = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
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
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
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
		ConditionType: ConditionTypeContainOnlyExactHost,
	}
	var rulePathCookie = ingressRouteRuleFile{
		RouteRuleFile: route_rule_conf.AdvancedRouteRuleFile{
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

func TestBfeRouteConfigBuilder_Build(t *testing.T) {
	testCases := map[string]interface{}{
		"single": map[string]interface{}{
			"annotation": map[string]interface{}{
				"cookie":       BfeRouteConfigBuilderBuildCaseDefinedAnnotation,
				"header":       BfeRouteConfigBuilderBuildCaseDefinedAnnotation,
				"load_balance": TestBfeRouteConfigBuilder_Build_CaseLoadBalance,
				"other":        TestBfeRouteConfigBuilder_Build_CaseOtherAnnotation,
			},
			"host": map[string]map[string]func(t *testing.T){
				"basic": {
					"wildcard": TestBfeRouteConfigBuilder_Build_CaseBasicRuleWildcardHost,
					"exact":    TestBfeRouteConfigBuilder_Build_CaseBasicRuleExactHost,
				},
				"advanced": {
					"wildcard": TestBfeRouteConfigBuilder_Build_CaseAdvancedRuleWildcardHost,
					"exact":    TestBfeRouteConfigBuilder_Build_CaseAdvancedRuleExactHost,
				},
			},
			"path": map[string]func(t *testing.T, name string){
				"prefix":                  RouteConfigBuilderBuildCasePrefixPath,
				"implementation_specific": RouteConfigBuilderBuildCasePrefixPath,
				"non_path_type":           RouteConfigBuilderBuildCasePrefixPath,
				"exact":                   RouteConfigBuilderBuildCaseExactPath,
			},
		},
		"multi": map[string]interface{}{
			"priority": map[string]map[string]func(t *testing.T){
				"path": {
					"exact_path_basic_exact_path_advanced":   TestBfeRouteConfigBuilder_Build_CasePriorityPath1,
					"exact_path_basic_path_advanced":         TestBfeRouteConfigBuilder_Build_CasePriorityPath2,
					"exact_path_basic_prefix_path_advanced":  TestBfeRouteConfigBuilder_Build_CasePriorityPath3,
					"prefix_path_basic_exact_path_advanced":  TestBfeRouteConfigBuilder_Build_CasePriorityPath4,
					"prefix_path_basic_prefix_path_advanced": TestBfeRouteConfigBuilder_Build_CasePriorityPath5,
				},
			},
			"conflict": TestBfeRouteConfigBuilder_Build_CaseConflict,
		},
	}

	traverseTestCases(t, testCases)
}

func TestBfeRouteConfigBuilder_Build2(t *testing.T) {

	b, err := routeConfigBuilderGenerator("normal")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	assert.Greater(t, len(*b.routeConf.routeTableFile.BasicRule), 0)
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func BfeRouteConfigBuilderBuildCaseDefinedAnnotation(t *testing.T, name string) {
	b, err := routeConfigBuilderGenerator("single/annotation/" + name)
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.BasicRule))
	assert.Greater(t, len(*b.routeConf.routeTableFile.ProductRule), 0)
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func TestBfeRouteConfigBuilder_Build_CaseOtherAnnotation(t *testing.T) {
	b, err := routeConfigBuilderGenerator("single/annotation/other")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	assert.Greater(t, len(*b.routeConf.routeTableFile.BasicRule), 0)
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func RouteConfigBuilderBuildCasePrefixPath(t *testing.T, name string) {
	b, err := routeConfigBuilderGenerator("single/path/" + name)
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	// path in basic rule with wildcard
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Greater(t, len(basicRules), 0)
	demoProductRules := basicRules[DemoHostExact]
	assert.Greater(t, len(demoProductRules), 0)
	paths := demoProductRules[0].Path
	assert.Greater(t, len(paths), 0)
	assert.Equal(t, paths[0], "/foo/*")
	// no advanced rule
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func RouteConfigBuilderBuildCaseExactPath(t *testing.T, name string) {
	b, err := routeConfigBuilderGenerator("single/path/" + name)
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	// host in basic rule
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Greater(t, len(basicRules), 0)
	demoProductRules := basicRules[DemoHostExact]
	assert.Greater(t, len(demoProductRules), 0)
	paths := demoProductRules[0].Path
	assert.Greater(t, len(paths), 0)
	assert.Equal(t, paths[0], "/foo")
	// no advanced rule
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func TestBfeRouteConfigBuilder_Build_CaseBasicRuleWildcardHost(t *testing.T) {
	b, err := routeConfigBuilderGenerator("single/host/basic/wildcard")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	// host in basic rule
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Greater(t, len(basicRules), 0)
	demoProductRules := basicRules[DemoHostWildcard]
	assert.Greater(t, len(demoProductRules), 0)
	hosts := demoProductRules[0].Hostname
	assert.Greater(t, len(hosts), 0)
	assert.Equal(t, hosts[0], DemoHostWildcard)
	// no advanced rule
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func TestBfeRouteConfigBuilder_Build_CaseBasicRuleExactHost(t *testing.T) {
	b, err := routeConfigBuilderGenerator("single/host/basic/exact")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	// host in basic rule
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Greater(t, len(basicRules), 0)
	demoProductRules := basicRules[DemoHostExact]
	assert.Greater(t, len(demoProductRules), 0)
	hosts := demoProductRules[0].Hostname
	assert.Greater(t, len(hosts), 0)
	assert.Equal(t, hosts[0], DemoHostExact)
	// no advanced rule
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func TestBfeRouteConfigBuilder_Build_CaseAdvancedRuleWildcardHost(t *testing.T) {
	b, err := routeConfigBuilderGenerator("single/host/advanced/wildcard")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	// no basic rule
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.BasicRule))
	// host in advanced rule
	advancedRule := *b.routeConf.routeTableFile.ProductRule
	assert.Greater(t, len(advancedRule), 0)
	demoProductRules := advancedRule[DemoHostWildcard]
	assert.Greater(t, len(demoProductRules), 0)
	cond := demoProductRules[0].Cond
	assert.NotNil(t, cond)
	assert.Contains(t, *cond, `req_host_suffix_in(".example.host.com")`)
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func TestBfeRouteConfigBuilder_Build_CaseAdvancedRuleExactHost(t *testing.T) {
	b, err := routeConfigBuilderGenerator("single/host/advanced/exact")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	// no basic rule
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.BasicRule))
	// host in advanced rule
	advancedRule := *b.routeConf.routeTableFile.ProductRule
	assert.Greater(t, len(advancedRule), 0)
	demoProductRules := advancedRule[DemoHostExact]
	assert.Greater(t, len(demoProductRules), 0)
	cond := demoProductRules[0].Cond
	assert.NotNil(t, cond)
	assert.Contains(t, *cond, `req_host_in("example.host.com")`)
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
}

func TestBfeRouteConfigBuilder_Build_CaseLoadBalance(t *testing.T) {
	b, _ := routeConfigBuilderGenerator("single/annotation/load_balance")

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))
	assert.Greater(t, len(*b.routeConf.routeTableFile.BasicRule), 0)
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
}

func TestBfeRouteConfigBuilder_Build_CaseConflict(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/conflict")
	assert.NotNil(t, err, "expect ingress conflicts")

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Greater(t, len(basicRules), 0)
	demoProductRules := basicRules["example.foo.com"]
	assert.Greater(t, len(demoProductRules), 0)
	// previous ingress rule
	assert.Contains(t, *demoProductRules[0].ClusterName, "service1")
	assert.Equal(t, 0, len(*b.routeConf.routeTableFile.ProductRule))
}

func TestBfeRouteConfigBuilder_Build_CasePriorityPath1(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/path/exact_path_basic_exact_path_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 1, len(basicRules)) // product count
	demoProductBasicRules := basicRules[DemoHostExact]
	assert.Equal(t, 2, len(demoProductBasicRules)) // basic rule count
	assert.Contains(t, *demoProductBasicRules[0].ClusterName, "service4")
	assert.Contains(t, demoProductBasicRules[0].Path, "/foo")
	assert.Equal(t, "ADVANCED_MODE", *demoProductBasicRules[1].ClusterName)
	assert.Contains(t, demoProductBasicRules[1].Path, "/bar")

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 1, len(advancedRules)) // product count
	demoProductAdvancedRules := advancedRules[DemoHostExact]
	assert.Equal(t, 2, len(demoProductAdvancedRules)) // advanced rule count
	assert.Contains(t, *demoProductAdvancedRules[0].ClusterName, "service1")
	assert.Contains(t, *demoProductAdvancedRules[1].ClusterName, "service3")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityPath2(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/path/exact_path_basic_prefix_path_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 1, len(basicRules)) // product count
	demoProductBasicRules := basicRules[DemoHostExact]
	assert.Equal(t, 1, len(demoProductBasicRules)) // basic rule count
	assert.Contains(t, demoProductBasicRules[0].Path, "/foo/bar")
	assert.Contains(t, *demoProductBasicRules[0].ClusterName, "service4")

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 1, len(advancedRules)) // product count
	demoProductAdvancedRules := advancedRules[DemoHostExact]
	assert.Equal(t, 3, len(demoProductAdvancedRules)) // advanced rule count
	assert.Contains(t, *demoProductAdvancedRules[0].ClusterName, "service3")
	assert.Contains(t, *demoProductAdvancedRules[1].ClusterName, "service2")
	assert.Contains(t, *demoProductAdvancedRules[2].ClusterName, "service1")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityPath3(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/path/exact_path_basic_path_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 1, len(basicRules)) // product count
	demoProductBasicRules := basicRules[DemoHostExact]
	assert.Equal(t, 1, len(demoProductBasicRules)) // basic rule count
	assert.Contains(t, demoProductBasicRules[0].Path, "/foo")
	assert.Equal(t, "ADVANCED_MODE", *demoProductBasicRules[0].ClusterName)

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 1, len(advancedRules)) // product count
	demoProductAdvancedRules := advancedRules[DemoHostExact]
	assert.Equal(t, 3, len(demoProductAdvancedRules)) // advanced rule count
	assert.Contains(t, *demoProductAdvancedRules[0].ClusterName, "service1")
	assert.Contains(t, *demoProductAdvancedRules[1].ClusterName, "service4")
	assert.Contains(t, *demoProductAdvancedRules[2].ClusterName, "service2")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityPath4(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/path/prefix_path_basic_exact_path_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 1, len(basicRules)) // product count
	demoProductBasicRules := basicRules[DemoHostExact]
	assert.Equal(t, 2, len(demoProductBasicRules)) // basic rule count
	assert.Contains(t, demoProductBasicRules[0].Path, "/foo/*")
	assert.Equal(t, "ADVANCED_MODE", *demoProductBasicRules[0].ClusterName)
	assert.Contains(t, demoProductBasicRules[1].Path, "/bar/*")
	assert.Equal(t, "ADVANCED_MODE", *demoProductBasicRules[1].ClusterName)

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 1, len(advancedRules)) // product count
	demoProductAdvancedRules := advancedRules[DemoHostExact]
	assert.Equal(t, 4, len(demoProductAdvancedRules)) // advanced rule count
	assert.Contains(t, *demoProductAdvancedRules[0].ClusterName, "service2")
	assert.Contains(t, *demoProductAdvancedRules[1].ClusterName, "service1")
	assert.Contains(t, *demoProductAdvancedRules[2].ClusterName, "service4")
	assert.Contains(t, *demoProductAdvancedRules[3].ClusterName, "service3")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityPath5(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/path/prefix_path_basic_prefix_path_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 1, len(basicRules)) // product count
	demoProductBasicRules := basicRules[DemoHostExact]
	assert.Equal(t, 3, len(demoProductBasicRules)) // basic rule count
	assert.Contains(t, demoProductBasicRules[0].Path, "/bar/baz/*")
	assert.Contains(t, *demoProductBasicRules[0].ClusterName, "service5")
	assert.Contains(t, demoProductBasicRules[1].Path, "/foo/*")
	assert.Equal(t, "ADVANCED_MODE", *demoProductBasicRules[1].ClusterName)
	assert.Contains(t, demoProductBasicRules[2].Path, "/bar/*")
	assert.Equal(t, "ADVANCED_MODE", *demoProductBasicRules[2].ClusterName)

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 1, len(advancedRules)) // product count
	demoProductAdvancedRules := advancedRules[DemoHostExact]
	assert.Equal(t, 4, len(demoProductAdvancedRules)) // advanced rule count
	assert.Contains(t, *demoProductAdvancedRules[0].ClusterName, "service1")
	assert.Contains(t, *demoProductAdvancedRules[1].ClusterName, "service2")
	assert.Contains(t, *demoProductAdvancedRules[2].ClusterName, "service3")
	assert.Contains(t, *demoProductAdvancedRules[3].ClusterName, "service4")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityHost1(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/host/exact_host_basic_exact_host_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 2, len(basicRules)) // product count
	assert.Equal(t, "ADVANCED_MODE", *basicRules["foo.host.com"][0].ClusterName)
	assert.Contains(t, *basicRules["bar.host.com"][0].ClusterName, "service3")

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 1, len(advancedRules))                 // product count
	assert.Equal(t, 2, len(advancedRules["foo.host.com"])) // advanced rule count
	assert.Contains(t, *advancedRules["foo.host.com"][0].ClusterName, "service1")
	assert.Contains(t, *advancedRules["foo.host.com"][1].ClusterName, "service2")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityHost2(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/host/exact_host_basic_wildcard_host_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 1, len(basicRules))                 // product count
	assert.Equal(t, 1, len(basicRules["foo.host.com"])) // basic rule count
	assert.Contains(t, *basicRules["foo.host.com"][0].ClusterName, "service4")

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 3, len(advancedRules)) // product count
	assert.Contains(t, *advancedRules["*.bar.foo.host.com"][0].ClusterName, "service1")
	assert.Contains(t, *advancedRules["*.foo.host.com"][0].ClusterName, "service2")
	assert.Contains(t, *advancedRules["*.host.com"][0].ClusterName, "service3")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityHost3(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/host/exact_host_basic_host_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 1, len(basicRules)) // product count
	assert.Equal(t, "ADVANCED_MODE", *basicRules["foo.host.com"][0].ClusterName)

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 2, len(advancedRules)) // product count
	assert.Contains(t, *advancedRules["foo.host.com"][0].ClusterName, "service1")
	assert.Contains(t, *advancedRules["foo.host.com"][1].ClusterName, "service3")
	assert.Contains(t, *advancedRules["*.host.com"][0].ClusterName, "service2")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityHost4(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/host/wildcard_host_basic_exact_host_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 3, len(basicRules)) // product count
	assert.Equal(t, "ADVANCED_MODE", *basicRules["*.foo.host.com"][0].ClusterName)
	assert.Contains(t, *basicRules["*.bar.host.com"][0].ClusterName, "service5")
	assert.Equal(t, "ADVANCED_MODE", *basicRules["*.baz.host.com"][0].ClusterName)

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 5, len(advancedRules)) // product count
	assert.Contains(t, *advancedRules["bar.foo.host.com"][0].ClusterName, "service1")
	assert.Contains(t, *advancedRules["*.foo.host.com"][0].ClusterName, "service4")
	assert.Contains(t, *advancedRules["bar.host.com"][0].ClusterName, "service2")
	assert.Contains(t, *advancedRules["foo.bar.baz.host.com"][0].ClusterName, "service3")
	assert.Contains(t, *advancedRules["*.baz.host.com"][0].ClusterName, "service6")
}

func TestBfeRouteConfigBuilder_Build_CasePriorityHost5(t *testing.T) {
	b, err := routeConfigBuilderGenerator("multi/priority/host/wildcard_host_basic_wildcard_host_advanced")
	assert.Nil(t, err, err)

	// invoke Build()
	if err := b.Build(); err != nil {
		t.Errorf("Build(): %s", err)
	}

	// verify
	t.Logf("routeTableFile: %s", jsonify(b.routeConf.routeTableFile))

	// 1 basic rule with ADVANCED_MODE
	basicRules := *b.routeConf.routeTableFile.BasicRule
	assert.Equal(t, 3, len(basicRules)) // product count
	assert.Equal(t, "ADVANCED_MODE", *basicRules["*.foo.host.com"][0].ClusterName)
	assert.Equal(t, "ADVANCED_MODE", *basicRules["*.bar.host.com"][0].ClusterName)
	assert.Contains(t, *basicRules["*.baz.host.com"][0].ClusterName, "service6")

	// 2 advanced rule
	advancedRules := *b.routeConf.routeTableFile.ProductRule
	assert.Equal(t, 4, len(advancedRules)) // product count
	assert.Contains(t, *advancedRules["*.bar.foo.host.com"][0].ClusterName, "service1")
	assert.Contains(t, *advancedRules["*.foo.host.com"][0].ClusterName, "service4")
	assert.Contains(t, *advancedRules["*.bar.host.com"][0].ClusterName, "service2")
	assert.Contains(t, *advancedRules["*.bar.host.com"][1].ClusterName, "service5")
	assert.Contains(t, *advancedRules["*.host.com"][0].ClusterName, "service3")
}

// routeConfigBuilderGenerator generate route config builder from file
// Params:
//		name: file name prefix
// Returns:
//		*BfeRouteConfigBuilder: builder generated by non-conflicting ingresses
// 		error: error for last conflict/wrong ingress
func routeConfigBuilderGenerator(name string) (*BfeRouteConfigBuilder, error) {

	// load ingress from file
	ingresses := loadIngress(name)

	// submit ingress to builder
	var submitErr error
	builder := NewBfeRouteConfigBuilder(nil, "0", nil)
	for _, ingress := range ingresses {
		err := builder.Submit(ingress)
		if err != nil {
			submitErr = err
		}
	}
	return builder, submitErr
}

func loadIngress(name string) []*networking.Ingress {
	fi, err := os.Open("./testdata/" + name + ".yaml")
	if err != nil {
		return nil
	}

	list := make([]*networking.Ingress, 0)
	decoder := yaml.NewYAMLOrJSONDecoder(fi, 4096)
	for {
		ingress := networking.Ingress{}
		err := decoder.Decode(&ingress)
		if err != nil {
			return list
		}
		list = append(list, &ingress)
	}
	return list
}

func jsonify(object interface{}) string {
	bytes, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return "marshal route table failed"
	}
	return string(bytes)
}

func traverseTestCases(t *testing.T, testCases interface{}) {
	v := reflect.ValueOf(testCases)
	for _, name := range v.MapKeys() {
		t.Run(name.String(), func(t *testing.T) {
			traverseNamedTestCases(t, name.String(), v.MapIndex(name).Interface())
		})
	}
}

func traverseNamedTestCases(t *testing.T, name string, testCases interface{}) {
	v := reflect.ValueOf(testCases)
	switch v.Kind() {
	case reflect.Func:
		switch v.Interface().(type) {
		case func(t *testing.T):
			v.Interface().(func(t *testing.T))(t)
			t.Logf("func(t *testing.T)")

		case func(t *testing.T, name string):
			v.Interface().(func(t *testing.T, name string))(t, name)
			t.Logf("func(t *testing.T, name string) \n")

		default:
			t.Errorf("unexpect func: %+v \n", v.Interface())
		}

	case reflect.Map:
		for _, name := range v.MapKeys() {
			t.Run(name.String(), func(t *testing.T) {
				traverseNamedTestCases(t, name.String(), v.MapIndex(name).Interface())
			})
		}

	default:
		t.Errorf("unexpect kind: %+v \n", v.Kind())
	}
}
