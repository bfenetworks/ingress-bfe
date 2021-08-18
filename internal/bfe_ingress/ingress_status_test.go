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
	"testing"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

func TestIngressStatusWriter_getErrorMsg(t *testing.T) {
	type fields struct {
		client *kubernetes_client.KubernetesClient
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name:   "normal",
			fields: fields{client: nil},
			args:   args{msg: "error msg"},
			want:   "{\"status\":\"error\",\"message\":\"error msg\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &IngressStatusWriter{
				client: tt.fields.client,
			}
			if got := w.getErrorMsg(tt.args.msg); got != tt.want {
				t.Errorf("IngressStatusWriter.getErrorMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIngressStatusWriter_getSuccessMsg(t *testing.T) {
	type fields struct {
		client *kubernetes_client.KubernetesClient
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name:   "normal",
			fields: fields{client: nil},
			args:   args{msg: "success msg"},
			want:   "{\"status\":\"success\",\"message\":\"success msg\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &IngressStatusWriter{
				client: tt.fields.client,
			}
			if got := w.getSuccessMsg(tt.args.msg); got != tt.want {
				t.Errorf("IngressStatusWriter.getSuccessMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
