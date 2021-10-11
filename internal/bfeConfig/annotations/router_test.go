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
	"testing"
)

func Test_cookieAnnotation_Build(t *testing.T) {

	tests := []struct {
		name   string
		fields map[string]string
		want   string
	}{
		{
			name: "cookie build",
			fields: map[string]string{
				CookieAnnotation: "Cookie-Test: Test",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test\", false)",
		},
		{
			name: "cookie build value with space",
			fields: map[string]string{
				CookieAnnotation: "Cookie-Test: Test ",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test\", false)",
		},
		{
			name: "cookie build key with space",
			fields: map[string]string{
				CookieAnnotation: "Cookie-Test : Test ",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test\", false)",
		},
		{
			name: "cookie build value with colon",
			fields: map[string]string{
				CookieAnnotation: "Cookie-Test: Test:123 ",
			},
			want: "req_cookie_value_in(\"Cookie-Test\", \"Test:123\", false)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := GetRouteExpression(tt.fields)
			if got != tt.want {
				fmt.Printf("%s, %s", got, tt.want)
				t.Errorf("GetRouteExpression() [%s] fail, got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_cookieAnnotation_Check(t *testing.T) {

	tests := []struct {
		name    string
		fields  map[string]string
		wantErr bool
	}{
		{
			name: "check normal",
			fields: map[string]string{
				CookieAnnotation: "Key: value",
			},
			wantErr: false,
		},
		{
			name: "check normal",
			fields: map[string]string{
				CookieAnnotation: "Key value",
			},
			wantErr: true,
		},
		{
			name: "check normal",
			fields: map[string]string{
				CookieAnnotation: "Key: value:123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GetRouteExpression(tt.fields); (err != nil) != tt.wantErr {
				t.Errorf("cookieAnnotation.Check() [%s] fail, error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func Test_headerAnnotation_Build(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]string
		want   string
	}{
		{
			name: "header build",
			fields: map[string]string{
				HeaderAnnotation: "Header-Test: Test",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test\", false)",
		},
		{
			name: "header build value with space",
			fields: map[string]string{
				HeaderAnnotation: "Header-Test: Test ",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test\", false)",
		},
		{
			name: "header build key with space",
			fields: map[string]string{
				HeaderAnnotation: "Header-Test : Test ",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test\", false)",
		},
		{
			name: "header build key with space",
			fields: map[string]string{
				HeaderAnnotation: "Header-Test: Test:123 ",
			},
			want: "req_header_value_in(\"Header-Test\", \"Test:123\", false)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetRouteExpression(tt.fields); got != tt.want {
				t.Errorf("headerAnnotation.Build() [%s] fail, go %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_headerAnnotation_Check(t *testing.T) {

	tests := []struct {
		name    string
		fields  map[string]string
		wantErr bool
	}{
		{
			name: "check normal",
			fields: map[string]string{
				HeaderAnnotation: "Key: value",
			},
			wantErr: false,
		},
		{
			name: "check normal",
			fields: map[string]string{
				HeaderAnnotation: "Key value",
			},
			wantErr: true,
		},
		{
			name: "check normal",
			fields: map[string]string{
				HeaderAnnotation: "Key: value:123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GetRouteExpression(tt.fields); (err != nil) != tt.wantErr {
				t.Errorf("headerAnnotation.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
