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

func TestBuildWeightAnnotation(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    Balance
		wantErr bool
	}{
		{
			name: "error key",
			args: args{
				key:   "error key",
				value: `{"service": {"service": 100}}`,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "normal",
			args: args{
				key:   WeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":67}}`,
			},
			want: Balance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
			},
			wantErr: false,
		},
		{
			name: "abnormal json",
			args: args{
				key:   WeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":"66"}}`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				key:   WeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":67}, "service3":{"service3-1": 33, "service3-2":33, "service3-3":33}}`,
			},
			want: Balance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
				"service3": map[string]int{"service3-1": 33, "service3-2": 33, "service3-3": 33},
			},
			wantErr: false,
		},
		{
			name: "sum zero",
			args: args{
				key:   WeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":66}, "service3":{"service3-1": 0, "service3-2":0, "service3-3":0}}`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "diff normal",
			args: args{
				key:   WeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":67}, "service3":{"service3-1": 1, "service3-2":2, "service3-3":100}}`,
			},
			want: Balance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
				"service3": map[string]int{"service3-1": 1, "service3-2": 2, "service3-3": 100},
			},
			wantErr: false,
		},
		{
			name: "diff normal 1 1 1",
			args: args{
				key:   WeightAnnotation,
				value: `{"service": {"service": 100}, "service2":{"service2-1": 33, "service2-2":67}, "service3":{"service3-1": 1, "service3-2":1, "service3-3":1}}`,
			},
			want: Balance{
				"service":  map[string]int{"service": 100},
				"service2": map[string]int{"service2-1": 33, "service2-2": 67},
				"service3": map[string]int{"service3-1": 1, "service3-2": 1, "service3-3": 1},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBalance(map[string]string{tt.args.key: tt.args.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance(), name=%s, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}
