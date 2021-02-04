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
	"fmt"
)

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
)

func GetSingleClusterName(namespace, serviceName string) string {
	return fmt.Sprintf("%s_%s", namespace, serviceName)
}
func GetMultiClusterName(namespace, ingressName, serviceKey string) string {
	return fmt.Sprintf("%s_%s_%s", namespace, ingressName, serviceKey)
}

func InitClusterGslb() *cluster_conf.GslbBasicConf {
	gslbConf := &cluster_conf.GslbBasicConf{}
	defaultCrossRetry := 0
	gslbConf.CrossRetry = &defaultCrossRetry

	defaultRetryMax := 2
	gslbConf.RetryMax = &defaultRetryMax

	defaultHashStrategy := cluster_conf.ClientIdOnly
	defaultHashHeader := "Cookie: bfe_userid"
	defaultSessionSticky := true

	gslbConf.HashConf = &cluster_conf.HashConf{
		HashStrategy:  &defaultHashStrategy,
		HashHeader:    &defaultHashHeader,
		SessionSticky: &defaultSessionSticky,
	}

	defaultBalMode := cluster_conf.BalanceModeWrr
	gslbConf.BalanceMode = &defaultBalMode
	return gslbConf
}
