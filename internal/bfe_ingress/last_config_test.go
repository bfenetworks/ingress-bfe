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

import "testing"

func Test_isConfigEqual(t *testing.T) {
	lastConfigMap = make(map[string]interface{})
	updateLastConfig("key1", 0)
	updateLastConfig("key2", "value")
	updateLastConfig("key3", struct{ key string }{key: "value"})
	type args struct {
		configName string
		newConfig  interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "equal",
			args: args{
				configName: "key1",
				newConfig:  0,
			},
			want: true,
		},
		{
			name: "equal",
			args: args{
				configName: "key2",
				newConfig:  "value",
			},
			want: true,
		},
		{
			name: "equal",
			args: args{
				configName: "key3",
				newConfig:  struct{ key string }{key: "value"},
			},
			want: true,
		},
		{
			name: "not equal",
			args: args{
				configName: "key4",
				newConfig:  0,
			},
			want: false,
		},
		{
			name: "not equal",
			args: args{
				configName: "key3",
				newConfig:  struct{ key string }{key: "value1"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isConfigEqual(tt.args.configName, tt.args.newConfig); got != tt.want {
				t.Errorf("isConfigEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
