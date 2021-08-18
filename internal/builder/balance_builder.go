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
	"fmt"
	"sort"
	"strconv"
	"strings"
)

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/gslb_conf"
	"github.com/bfenetworks/bfe/bfe_util"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
	"github.com/bfenetworks/ingress-bfe/internal/utils"
)

const (
	ConfigNameBalanceConf = "gslb_data_conf"

	GslbData         = "cluster_conf/gslb.data"
	ClusterTableData = "cluster_conf/cluster_table.data"
)

type BfeBalanceConf struct {
	gslbConf         *gslb_conf.GslbConf
	clusterTableConf *cluster_table_conf.ClusterTableConf
}

type ingressSubCluster struct {
	name   string
	port   string
	weight int
}

type ingressSubClusters struct {
	subClusters []*ingressSubCluster
	refCount    int
}

type ingressClusters map[string]*ingressSubClusters

type BfeBalanceConfigBuilder struct {
	client   *kubernetes_client.KubernetesClient
	reloader *Reloader
	version  string
	hostName string

	balanceConf BfeBalanceConf

	clusters ingressClusters
}

func NewBfeBalanceConfigBuilder(client *kubernetes_client.KubernetesClient, version string, reloader *Reloader) *BfeBalanceConfigBuilder {
	c := &BfeBalanceConfigBuilder{}
	c.client = client
	c.version = version
	c.reloader = reloader
	c.hostName = "bfe-ingress-controller"
	c.clusters = make(ingressClusters)
	return c
}

func (c *BfeBalanceConfigBuilder) submitClusters(clusterName string, subClusterList []*ingressSubCluster) error {
	if _, ok := c.clusters[clusterName]; !ok {
		c.clusters[clusterName] = &ingressSubClusters{
			subClusters: subClusterList,
			refCount:    1,
		}
	} else {
		c.clusters[clusterName].refCount++
	}
	return nil
}

func (c *BfeBalanceConfigBuilder) rollbackClusters(clusterName string) error {
	if _, ok := c.clusters[clusterName]; !ok {
		return nil
	} else {
		c.clusters[clusterName].refCount--
		if c.clusters[clusterName].refCount == 0 {
			delete(c.clusters, clusterName)
		}
	}
	return nil
}

func (c *BfeBalanceConfigBuilder) Submit(ingress *networking.Ingress) error {
	var err error

	// parse load-balance parameters from annotation
	var balance LoadBalance
	for key, value := range ingress.Annotations {
		if key == LoadBalanceWeightAnnotation {
			balance, err = BuildLoadBalanceAnnotation(key, value)
			if err != nil {
				return err
			}
			break
		}
	}

	type cacheItem struct {
		clusterName string
		subCluster  []*ingressSubCluster
	}
	var cache = make([]cacheItem, 0)

	for _, rule := range ingress.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			if !balance.ContainService(p.Backend.ServiceName) {
				clusterName := GetSingleClusterName(ingress.Namespace, p.Backend.ServiceName)
				subClusterName := p.Backend.ServiceName

				eps, err := c.client.GetEndpoints(ingress.Namespace, p.Backend.ServiceName)
				if err != nil {
					return fmt.Errorf("[%s/Services/%s] get endpoints error: %s",
						ingress.Namespace, p.Backend.ServiceName, err.Error())
				}
				if len(eps.Subsets) == 0 {
					return fmt.Errorf("[%s/Services/%s] has no backend", ingress.Namespace, p.Backend.ServiceName)
				}

				subClusterObj := ingressSubCluster{
					name:   subClusterName,
					port:   p.Backend.ServicePort.StrVal,
					weight: 100,
				}
				cache = append(cache, cacheItem{clusterName, []*ingressSubCluster{&subClusterObj}})

			} else {
				clusterName := GetMultiClusterName(ingress.Namespace, ingress.Name, p.Backend.ServiceName)
				subClusters, _ := balance.GetService(p.Backend.ServiceName)

				ingressSubList := make([]*ingressSubCluster, 0)
				for subClusterName, weight := range subClusters {
					eps, err := c.client.GetEndpoints(ingress.Namespace, subClusterName)
					if err != nil {
						return fmt.Errorf("[%s/Services/%s] get endpoints error: %s",
							ingress.Namespace, p.Backend.ServiceName, err.Error())
					}
					if len(eps.Subsets) == 0 {
						return fmt.Errorf("[%s/Services/%s] has no backend", ingress.Namespace, p.Backend.ServiceName)
					}
					if weight < 0 {
						return fmt.Errorf("[%s/Services/%s] invalid weight %d, less than zero",
							ingress.Namespace, subClusterName, weight)
					}
					subClusterObj := ingressSubCluster{
						name:   subClusterName,
						port:   p.Backend.ServicePort.StrVal,
						weight: weight,
					}
					ingressSubList = append(ingressSubList, &subClusterObj)
				}
				cache = append(cache, cacheItem{clusterName, ingressSubList})
			}
		}
	}
	for _, item := range cache {
		c.submitClusters(item.clusterName, item.subCluster)
	}
	return nil
}

func (c *BfeBalanceConfigBuilder) Rollback(ingress *networking.Ingress) error {
	var balance LoadBalance
	var err error
	for key, value := range ingress.Annotations {
		if key == LoadBalanceWeightAnnotation {
			balance, err = BuildLoadBalanceAnnotation(key, value)
			if err != nil {
				return err
			}
			break
		}
	}

	for _, rule := range ingress.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			var clusterName string
			if !balance.ContainService(p.Backend.ServiceName) {
				clusterName = GetSingleClusterName(ingress.Namespace, p.Backend.ServiceName)
			} else {
				clusterName = GetMultiClusterName(ingress.Namespace, ingress.Name, p.Backend.ServiceName)
			}
			err := c.rollbackClusters(clusterName)
			if err != nil {
				return fmt.Errorf("Rollback ingress error: %s", err.Error())
			}
		}
	}
	return nil
}

func (c *BfeBalanceConfigBuilder) Build() error {
	var gslbConf gslb_conf.GslbConf
	var clusterTableConf cluster_table_conf.ClusterTableConf
	c.balanceConf = BfeBalanceConf{
		clusterTableConf: &clusterTableConf,
		gslbConf:         &gslbConf,
	}

	c.balanceConf.clusterTableConf.Version = &c.version
	c.balanceConf.gslbConf.Ts = &c.version
	c.balanceConf.gslbConf.Hostname = &c.hostName

	allCluster, err := c.buildAllClusterBackend()
	if err != nil {
		return err
	}
	c.balanceConf.clusterTableConf.Config = &allCluster

	gslbCluster, err := c.buildGslbConf()
	if err != nil {
		return err
	}
	c.balanceConf.gslbConf.Clusters = &gslbCluster

	return nil
}

func (c *BfeBalanceConfigBuilder) buildGslbConf() (gslb_conf.GslbClustersConf, error) {
	gslbClustersConf := make(gslb_conf.GslbClustersConf)

	for key, subClusters := range c.clusters {
		gslbClusterConf := make(gslb_conf.GslbClusterConf)
		for _, subCluster := range (*subClusters).subClusters {
			gslbClusterConf[subCluster.name] = subCluster.weight
		}
		gslbClustersConf[key] = gslbClusterConf
	}

	return gslbClustersConf, nil
}

func (c *BfeBalanceConfigBuilder) buildAllClusterBackend() (cluster_table_conf.AllClusterBackend, error) {
	allClusterBackend := make(cluster_table_conf.AllClusterBackend)

	for key, subClusters := range c.clusters {
		namespace := strings.Split(key, "_")[0]
		for _, subCluster := range (*subClusters).subClusters {
			clusterBackend, err := c.buildClusterBackend(namespace, (*subCluster).name, (*subCluster).port)
			if err != nil {
				return allClusterBackend, err
			}
			if _, ok := allClusterBackend[key]; !ok {
				allClusterBackend[key] = make(cluster_table_conf.ClusterBackend)
			}
			for subClusterName, val := range clusterBackend {
				allClusterBackend[key][subClusterName] = val
			}
		}
	}
	return allClusterBackend, nil

}

func (c *BfeBalanceConfigBuilder) buildClusterBackend(namespace, serviceName string, port string) (cluster_table_conf.ClusterBackend, error) {
	var clusterBackend cluster_table_conf.ClusterBackend

	eps, err := c.client.GetEndpoints(namespace, serviceName)

	if err != nil {
		return clusterBackend, err
	}
	subClusterBackend := c.buildSubClusterBackend(eps, port)
	clusterBackend = make(cluster_table_conf.ClusterBackend)

	sort.Slice(subClusterBackend, func(i, j int) bool {
		return *subClusterBackend[i].Name > *subClusterBackend[j].Name
	})

	if len(subClusterBackend) == 0 {
		return clusterBackend, fmt.Errorf("[%s/Services/%s] has no endpoints", namespace, serviceName)
	}

	clusterBackend[serviceName] = subClusterBackend
	return clusterBackend, nil

}

func (c *BfeBalanceConfigBuilder) buildSubClusterBackend(eps *core.Endpoints, port string) cluster_table_conf.SubClusterBackend {
	var subClusterBackend cluster_table_conf.SubClusterBackend
	defaultWeight := 1
	for _, subsets := range eps.Subsets {
		for _, addr := range subsets.Addresses {
			if port != "" {
				name := fmt.Sprintf("%s:%s", addr.IP, port)
				ip := addr.IP
				portVal, _ := strconv.Atoi(port)
				backendConf := cluster_table_conf.BackendConf{
					Name:   &name,
					Addr:   &ip,
					Port:   &portVal,
					Weight: &defaultWeight,
				}
				subClusterBackend = append(subClusterBackend, &backendConf)
			} else {
				for _, setPort := range subsets.Ports {
					name := fmt.Sprintf("%s:%d", addr.IP, setPort.Port)
					portVal := int(setPort.Port)
					ip := addr.IP
					backendConf := cluster_table_conf.BackendConf{
						Name:   &name,
						Addr:   &ip,
						Port:   &portVal,
						Weight: &defaultWeight,
					}
					subClusterBackend = append(subClusterBackend, &backendConf)
				}
			}
		}
	}
	return subClusterBackend
}

func (c *BfeBalanceConfigBuilder) Dump() error {
	err := bfe_util.DumpJson(c.balanceConf.gslbConf, utils.ConfigPath+GslbData, utils.FilePerm)
	if err != nil {
		return fmt.Errorf("dump gslb.data error: %v", err)
	}

	err = bfe_util.DumpJson(c.balanceConf.clusterTableConf, utils.ConfigPath+ClusterTableData, utils.FilePerm)
	if err != nil {
		return fmt.Errorf("dump cluster_table.data error: %v", err)
	}

	return nil
}

func (c *BfeBalanceConfigBuilder) Reload() error {
	return c.reloader.DoReload(c.balanceConf, ConfigNameBalanceConf)
}

func GetSingleClusterName(namespace, serviceName string) string {
	return fmt.Sprintf("%s_%s", namespace, serviceName)
}
func GetMultiClusterName(namespace, ingressName, serviceKey string) string {
	return fmt.Sprintf("%s_%s_%s", namespace, ingressName, serviceKey)
}
