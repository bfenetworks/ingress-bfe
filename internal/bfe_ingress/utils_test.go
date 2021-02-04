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

	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
)

func TestGetSingleClusterName(t *testing.T) {
	type args struct {
		namespace   string
		serviceName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				namespace:   "poc",
				serviceName: "service",
			},
			want: "poc_service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSingleClusterName(tt.args.namespace, tt.args.serviceName); got != tt.want {
				t.Errorf("GetSingleClusterName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMultiClusterName(t *testing.T) {
	type args struct {
		namespace   string
		ingressName string
		serviceKey  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				namespace:   "poc",
				ingressName: "ingress-name",
				serviceKey:  "service-name",
			},
			want: "poc_ingress-name_service-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMultiClusterName(tt.args.namespace, tt.args.ingressName, tt.args.serviceKey); got != tt.want {
				t.Errorf("GetMultiClusterName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitClusterGslb(t *testing.T) {
	defaultCrossRetry := 0
	defaultRetryMax := 2
	defaultHashStrategy := cluster_conf.ClientIdOnly
	defaultHashHeader := "Cookie: bfe_userid"
	defaultSessionSticky := true
	defaultBalMode := cluster_conf.BalanceModeWrr
	tests := []struct {
		name string
		want *cluster_conf.GslbBasicConf
	}{
		{
			name: "normal",
			want: &cluster_conf.GslbBasicConf{
				CrossRetry: &defaultCrossRetry,
				RetryMax:   &defaultRetryMax,
				HashConf: &cluster_conf.HashConf{
					HashStrategy:  &defaultHashStrategy,
					HashHeader:    &defaultHashHeader,
					SessionSticky: &defaultSessionSticky,
				},
				BalanceMode: &defaultBalMode,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitClusterGslb(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitClusterGslb() = %v, want %v", got, tt.want)
			}
		})
	}
}
