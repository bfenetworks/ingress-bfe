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
	"reflect"
	"testing"
)

func TestGetRewriteAction(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    RewriteAction
		wantErr bool
	}{
		{
			name: "error rewrite annotation",
			args: args{
				key:   "aaaa",
				value: `[{"params": "bar.com", "order": 1, "when": "AfterLocation"}]`,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "error rewrite param",
			args: args{
				key:   RewriteURLHostSetAnnotation,
				value: `aaaaa`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing param",
			args: args{
				key:   RewriteURLHostSetAnnotation,
				value: `[{"order": 1, "when": "AfterLocation"}]`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing when",
			args: args{
				key:   RewriteURLHostSetAnnotation,
				value: `[{"params": "bar.com", "order": 1}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"HOST_SET": &RewriteActionParam{
						Params:   []string{"bar.com"},
						Order:    1,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing order",
			args: args{
				key:   RewriteURLHostSetAnnotation,
				value: `[{"params": "bar.com", "when": "AfterLocation"}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"HOST_SET": &RewriteActionParam{
						Params:   []string{"bar.com"},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal host_set",
			args: args{
				key:   RewriteURLHostSetAnnotation,
				value: `[{"params": "bar.com", "order": 1, "when": "AfterLocation"}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"HOST_SET": &RewriteActionParam{
						Params:   []string{"bar.com"},
						Order:    1,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error host_set",
			args: args{
				key:   RewriteURLHostSetAnnotation,
				value: `[{"params": "bar.com.*"}]`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal host_set_from_path",
			args: args{
				key:   RewriteURLHostFromPathAnnotation,
				value: `[{"params": true}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"HOST_SET_FROM_PATH_PREFIX": &RewriteActionParam{
						Params:   []string{},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error host_set_from_path",
			args: args{
				key:   RewriteURLHostFromPathAnnotation,
				value: `[{"params": false }]`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal path_set",
			args: args{
				key:   RewriteURLPathSetAnnotation,
				value: `[{"params": "/bar"}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"PATH_SET": &RewriteActionParam{
						Params:   []string{"/bar"},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error wildcard path",
			args: args{
				key:   RewriteURLPathSetAnnotation,
				value: `[{"params": "/bar*"}]`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "error path_set",
			args: args{
				key:   RewriteURLPathSetAnnotation,
				value: `[{"params": ""}]`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal path_prefix_add",
			args: args{
				key:   RewriteURLPathPrefixAddAnnotation,
				value: `[{"params": "/bar/"}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"PATH_PREFIX_ADD": &RewriteActionParam{
						Params:   []string{"/bar/"},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal, prefix_path not end with /",
			args: args{
				key:   RewriteURLPathPrefixAddAnnotation,
				value: `[{"params": "/bar"}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"PATH_PREFIX_ADD": &RewriteActionParam{
						Params:   []string{"/bar/"},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal query_add",
			args: args{
				key:   RewriteURLQueryAddAnnotation,
				value: `[{"params": {"name": "user"}}]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"QUERY_ADD": &RewriteActionParam{
						Params:   []string{"name", "user"},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error, empty param",
			args: args{
				key:   RewriteURLQueryAddAnnotation,
				value: `[{"params": {}}]`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal query_del",
			args: args{
				key:   RewriteURLQueryDeleteAnnotation,
				value: `[{"params": ["name", "user"] }]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"QUERY_DEL": &RewriteActionParam{
						Params:   []string{"name", "user"},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error query_del",
			args: args{
				key:   RewriteURLQueryDeleteAnnotation,
				value: `[{"params": [] }]`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal query_del_all_except",
			args: args{
				key:   RewriteURLQueryDeleteAllExceptAnnotation,
				value: `[{"params": "name" }]`,
			},
			want: RewriteAction{
				DefaultCallBackPoint: {
					"QUERY_DEL_ALL_EXCEPT": &RewriteActionParam{
						Params:   []string{"name"},
						Order:    0,
						Callback: DefaultCallBackPoint,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRewriteAction(map[string]string{tt.args.key: tt.args.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPathStripPrefix(), name=%s, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPathStripPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPathStripPrefix(t *testing.T) {
	type args struct {
		path   string
		length string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "invalid length",
			args: args{
				path:   "/bar/foo",
				length: "aaa",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "negative length",
			args: args{
				path:   "/bar/foo",
				length: "-1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				path:   "/bar/foo",
				length: "1",
			},
			want:    "/bar",
			wantErr: false,
		},
		{
			name: "root path",
			args: args{
				path:   "/",
				length: "1",
			},
			want:    "/",
			wantErr: false,
		},
		{
			name: "wildcard path",
			args: args{
				path:   "/bar*",
				length: "1",
			},
			want:    "/bar",
			wantErr: false,
		},
		{
			name: "strip length greater than path segments",
			args: args{
				path:   "/bar",
				length: "2",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPathStripPrefix(tt.args.path, tt.args.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPathStripPrefix(), name=%s, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetPathStripPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckAllowedCallBack(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				value: DefaultCallBackPoint,
			},
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				value: "BeforeForward",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAllowedCallBack(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPathStripPrefix(), name=%s, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
		})
	}
}
