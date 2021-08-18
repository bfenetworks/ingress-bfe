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
	"reflect"
	"testing"
)

func Test_cookieAnnotation_Build(t *testing.T) {
	type fields struct {
		annotationStr string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "cookie build",
			fields: fields{
				annotationStr: "Cookie-Test: Test",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test\", false)",
		},
		{
			name: "cookie build value with space",
			fields: fields{
				annotationStr: "Cookie-Test: Test ",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test\", false)",
		},
		{
			name: "cookie build key with space",
			fields: fields{
				annotationStr: "Cookie-Test : Test ",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test\", false)",
		},
		{
			name: "cookie build key with space",
			fields: fields{
				annotationStr: "Cookie-Test: Test:123 ",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test:123\", false)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := &cookieAnnotation{
				annotationStr: tt.fields.annotationStr,
			}
			if got := cookie.Build(); got != tt.want {
				t.Errorf("cookieAnnotation.Build() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cookieAnnotation_Check(t *testing.T) {
	type fields struct {
		annotationStr string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "check normal",
			fields: fields{
				annotationStr: "Key: value",
			},
			wantErr: false,
		},
		{
			name: "check normal",
			fields: fields{
				annotationStr: "Key value",
			},
			wantErr: true,
		},
		{
			name: "check normal",
			fields: fields{
				annotationStr: "Key: value:123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := &cookieAnnotation{
				annotationStr: tt.fields.annotationStr,
			}
			if err := cookie.Check(); (err != nil) != tt.wantErr {
				t.Errorf("cookieAnnotation.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_headerAnnotation_Build(t *testing.T) {
	type fields struct {
		annotationStr string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "header build",
			fields: fields{
				annotationStr: "Header-Test: Test",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test\", false)",
		},
		{
			name: "header build value with space",
			fields: fields{
				annotationStr: "Header-Test: Test ",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test\", false)",
		},
		{
			name: "header build key with space",
			fields: fields{
				annotationStr: "Header-Test : Test ",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test\", false)",
		},
		{
			name: "header build key with space",
			fields: fields{
				annotationStr: "Header-Test: Test:123 ",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test:123\", false)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &headerAnnotation{
				annotationStr: tt.fields.annotationStr,
			}
			if got := header.Build(); got != tt.want {
				t.Errorf("headerAnnotation.Build() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_headerAnnotation_Check(t *testing.T) {
	type fields struct {
		annotationStr string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "check normal",
			fields: fields{
				annotationStr: "Key: value",
			},
			wantErr: false,
		},
		{
			name: "check normal",
			fields: fields{
				annotationStr: "Key value",
			},
			wantErr: true,
		},
		{
			name: "check normal",
			fields: fields{
				annotationStr: "Key: value:123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &headerAnnotation{
				annotationStr: tt.fields.annotationStr,
			}
			if err := header.Check(); (err != nil) != tt.wantErr {
				t.Errorf("headerAnnotation.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildBfeAnnotation(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    BfeAnnotation
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "build header",
			args: args{
				key:   "bfe.ingress.kubernetes.io/router.header",
				value: "Header: Value",
			},
			want:    &headerAnnotation{annotationStr: "Header: Value"},
			wantErr: false,
		},
		{
			name: "build cookie",
			args: args{
				key:   "bfe.ingress.kubernetes.io/router.cookie",
				value: "Cookie: Value",
			},
			want:    &cookieAnnotation{annotationStr: "Cookie: Value"},
			wantErr: false,
		},
		{
			name: "build err",
			args: args{
				key:   "bfe.ingress.kubernetes.io/router.err",
				value: "Header: Value",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildBfeAnnotation(tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildBfeAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildBfeAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortAnnotations(t *testing.T) {
	type args struct {
		annotationConds []BfeAnnotation
	}
	tests := []struct {
		name string
		args args
		want []BfeAnnotation
	}{
		// TODO: Add test cases.
		{
			name: "sort normal",
			args: args{
				annotationConds: []BfeAnnotation{
					&headerAnnotation{},
					&cookieAnnotation{},
				},
			},
			want: []BfeAnnotation{
				&cookieAnnotation{},
				&headerAnnotation{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortAnnotations(tt.args.annotationConds)
		})
	}
}

func TestBuildLoadBalanceAnnotation(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    LoadBalance
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "error key",
			args: args{
				key:   "error key",
				value: `{"service": {"service": 100}}`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				key:   LoadBalanceWeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}}`,
			},
			want: LoadBalance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
			},
			wantErr: false,
		},
		{
			name: "abnormal json",
			args: args{
				key:   LoadBalanceWeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":"66"}}`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				key:   LoadBalanceWeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}, "service3":{"service3-1": 33, "service3-2":33, "service3-3":33}}`,
			},
			want: LoadBalance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
				"service3": map[string]int{"service3-1": 33, "service3-2": 33, "service3-3": 34},
			},
			wantErr: false,
		},
		{
			name: "sum zero",
			args: args{
				key:   LoadBalanceWeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}, "service3":{"service3-1": 0, "service3-2":0, "service3-3":0}}`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "diff normal",
			args: args{
				key:   LoadBalanceWeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}, "service3":{"service3-1": 1, "service3-2":2, "service3-3":100000}}`,
			},
			want: LoadBalance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
				"service3": map[string]int{"service3-1": 0, "service3-2": 0, "service3-3": 100},
			},
			wantErr: false,
		},
		{
			name: "diff normal 1 1 1",
			args: args{
				key:   LoadBalanceWeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}, "service3":{"service3-1": 1, "service3-2":1, "service3-3":1}}`,
			},
			want: LoadBalance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
				"service3": map[string]int{"service3-1": 33, "service3-2": 33, "service3-3": 34},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildLoadBalanceAnnotation(tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildLoadBalanceAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildLoadBalanceAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadBalance_GetService(t *testing.T) {
	val := `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}, "service3":{"service3-1": 33, "service3-2":33, "service3-3":33}}`
	l, _ := BuildLoadBalanceAnnotation(LoadBalanceWeightAnnotation, val)
	type args struct {
		serviceName string
	}
	tests := []struct {
		name    string
		l       *LoadBalance
		args    args
		want    ServicesWeight
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "normal",
			l:    &l,
			args: args{
				serviceName: "service",
			},
			want: map[string]int{
				"service": 100,
			},
			wantErr: false,
		},
		{
			name: "normal",
			l:    &l,
			args: args{
				serviceName: "service3",
			},
			want: map[string]int{
				"service3-1": 33,
				"service3-2": 33,
				"service3-3": 34,
			},
			wantErr: false,
		},
		{
			name: "normal",
			l:    &l,
			args: args{
				serviceName: "service4",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.GetService(tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadBalance.GetService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadBalance.GetService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadBalance_ContainService(t *testing.T) {
	val := `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}, "service3":{"service3-1": 33, "service3-2":33, "service3-3":33}}`
	l, _ := BuildLoadBalanceAnnotation(LoadBalanceWeightAnnotation, val)
	type args struct {
		service string
	}
	tests := []struct {
		name string
		l    *LoadBalance
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "normal",
			l:    &l,
			args: args{
				service: "service",
			},
			want: true,
		},
		{
			name: "normal",
			l:    &l,
			args: args{
				service: "service4",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.ContainService(tt.args.service); got != tt.want {
				t.Errorf("LoadBalance.ContainService() = %v, want %v", got, tt.want)
			}
		})
	}
}
