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

package filter

import (
	"testing"

	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
)

func Test_matchIngressClass(t *testing.T) {
	type args struct {
		targetCls *string
		testCls   v1.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match specific",
			args: args{
				targetCls: pointer.String("bfe"),
				testCls: &netv1.IngressClass{
					ObjectMeta: v1.ObjectMeta{
						Name: "bfe",
					},
				},
			},
			want: true,
		},
		{
			name: "mismatch specific",
			args: args{
				targetCls: pointer.String("bfe"),
				testCls: &netv1.IngressClass{
					ObjectMeta: v1.ObjectMeta{
						Name: "other",
					},
				},
			},
			want: false,
		},
		{
			name: "match default",
			args: args{
				targetCls: nil,
				testCls: &netv1.IngressClass{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							annotations.IsDefaultIngressClass: "True",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "mismatch default",
			args: args{
				targetCls: nil,
				testCls:   &netv1.IngressClass{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchIngressClass(tt.args.targetCls, tt.args.testCls); got != tt.want {
				t.Errorf("matchIngressClass() = %v, want %v", got, tt.want)
			}
		})
	}
}
