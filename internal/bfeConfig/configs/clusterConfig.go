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
package configs

import (
	"fmt"

	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/gslb_conf"
	"github.com/jwangsadinata/go-multimap/setmultimap"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/configs/log"
	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/util"
	"github.com/bfenetworks/ingress-bfe/internal/option"
)

const (
	ConfigNameclusterConf = "gslb_data_conf"

	GslbData         = "cluster_conf/gslb.data"
	ClusterTableData = "cluster_conf/cluster_table.data"
)

var (
	defaultWeight = 10
)

type ClusterConfig struct {
	gslbVersion         string // current active version in bfe
	clusterTableVersion string

	ingress2Cluster *setmultimap.MultiMap
	service2Cluster *setmultimap.MultiMap

	gslbConf         gslb_conf.GslbConf
	clusterTableConf cluster_table_conf.ClusterTableConf
}

func NewClusterConfig(version string) *ClusterConfig {
	hostname := "bfe-ingress-controller"

	gslbCluster := make(gslb_conf.GslbClustersConf)
	clusterBackend := make(cluster_table_conf.AllClusterBackend)

	return &ClusterConfig{
		ingress2Cluster: setmultimap.New(),
		service2Cluster: setmultimap.New(),
		gslbConf: gslb_conf.GslbConf{
			Clusters: &gslbCluster,
			Hostname: &hostname,
			Ts:       &version,
		},
		clusterTableConf: cluster_table_conf.ClusterTableConf{
			Version: &version,
			Config:  &clusterBackend,
		},
	}
}

func (c *ClusterConfig) setVersion() {
	version := util.NewVersion()

	c.gslbConf.Ts = &version
	c.clusterTableConf.Version = &version
}

func (c *ClusterConfig) UpdateIngress(ingress *netv1.Ingress, services map[string]*corev1.Service, endpoints map[string]*corev1.Endpoints) error {
	if len(ingress.Spec.Rules) == 0 {
		return nil
	}

	balance, _ := annotations.GetBalance(ingress.Annotations)

	ingressName := util.NamespacedName(ingress.Namespace, ingress.Name)
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			// create cluster && subcluster for each Service
			clusterName := util.ClusterName(ingressName, path.Backend.Service)

			// cluster config
			(*c.clusterTableConf.Config)[clusterName] = c.newClusterBackend(ingress.Namespace, path.Backend.Service, balance, services, endpoints)

			// gslb config
			(*c.gslbConf.Clusters)[clusterName] = c.newGslbClusterConf(ingress.Namespace, path.Backend.Service.Name, balance)

			// put into map
			c.ingress2Cluster.Put(ingressName, clusterName)
			for service := range (*c.gslbConf.Clusters)[clusterName] {
				c.service2Cluster.Put(service, clusterName)
			}
		}
	}

	if len(option.Opts.Ingress.DefaultBackend) > 0 {
		c.addDefautBackend(endpoints[option.Opts.Ingress.DefaultBackend])
	}

	if err := cluster_table_conf.ClusterTableConfCheck(c.clusterTableConf); err != nil {
		c.DeleteIngress(ingress.Namespace, ingress.Name)
		return err
	}

	c.setVersion()
	return nil
}

func (c *ClusterConfig) addDefautBackend(ep *corev1.Endpoints) {
	if ep == nil {
		return
	}

	// already exist
	if _, ok := (*c.clusterTableConf.Config)[util.DefaultClusterName()]; ok {
		return
	}

	instanceList := c.newSubClusterBackend(ep, intstr.IntOrString{})
	if len(instanceList) == 0 {
		return
	}

	serviceName := option.Opts.Ingress.DefaultBackend

	subCluster := make(cluster_table_conf.ClusterBackend)
	subCluster[serviceName] = instanceList
	(*c.clusterTableConf.Config)[util.DefaultClusterName()] = subCluster

	gslbConf := make(gslb_conf.GslbClusterConf)
	gslbConf[serviceName] = defaultWeight
	(*c.gslbConf.Clusters)[util.DefaultClusterName()] = gslbConf

	c.service2Cluster.Put(serviceName, util.DefaultClusterName())
}

func (c *ClusterConfig) DeleteIngress(namespace, name string) {
	ingressName := util.NamespacedName(namespace, name)
	clusters, ok := c.ingress2Cluster.Get(ingressName)
	if !ok {
		return
	}

	for _, cluster := range clusters {
		clusterName := cluster.(string)

		for serviceName := range (*c.clusterTableConf.Config)[clusterName] {
			c.service2Cluster.Remove(serviceName, clusterName)
		}

		delete(*c.clusterTableConf.Config, clusterName)
		delete(*c.gslbConf.Clusters, clusterName)
	}
	c.ingress2Cluster.RemoveAll(ingressName)

	// if no ingress exist, remove default backend config
	if len(option.Opts.Ingress.DefaultBackend) > 0 && c.ingress2Cluster.Empty() {
		c.delDefautBackend()
	}

	c.setVersion()
}

func (c *ClusterConfig) delDefautBackend() {
	c.service2Cluster.Remove(option.Opts.Ingress.DefaultBackend, util.DefaultClusterName())
	delete(*c.clusterTableConf.Config, util.DefaultClusterName())
	delete(*c.gslbConf.Clusters, util.DefaultClusterName())
}

// newClusterBackend makes cluster_table_conf.ClusterBackend configuration
func (c *ClusterConfig) newClusterBackend(namespace string, backend *netv1.IngressServiceBackend, balance annotations.Balance, services map[string]*corev1.Service, endpoints map[string]*corev1.Endpoints) cluster_table_conf.ClusterBackend {

	subClusters := make(cluster_table_conf.ClusterBackend)

	if backend == nil {
		return subClusters
	}
	// check whether service exist in balance annotation
	weights, ok := balance[backend.Name]
	if !ok {
		serviceName := util.NamespacedName(namespace, backend.Name)
		port := getTargetPort(backend.Port, services[serviceName])
		subClusters[serviceName] = c.newSubClusterBackend(endpoints[serviceName], port)
		return subClusters
	}

	for name := range weights {
		serviceName := util.NamespacedName(namespace, name)
		port := getTargetPort(backend.Port, services[serviceName])
		subClusters[serviceName] = c.newSubClusterBackend(endpoints[serviceName], port)
	}

	return subClusters
}

// newSubClusterBackend converts k8s service to bfe subCluster/instanceList
func (c *ClusterConfig) newSubClusterBackend(ep *corev1.Endpoints, port intstr.IntOrString) cluster_table_conf.SubClusterBackend {
	if ep == nil {
		return nil
	}
	instanceList := make([]*cluster_table_conf.BackendConf, 0)

	// if no port is specified, use the first port in endpoints
	if port.IntVal == 0 && len(port.StrVal) == 0 {
		if len(ep.Subsets) > 0 && len(ep.Subsets[0].Ports) > 0 && len(ep.Subsets[0].Addresses) > 0 {
			instanceList = append(instanceList, newBackendConf(ep.Subsets[0].Addresses[0].IP, int(ep.Subsets[0].Ports[0].Port), defaultWeight))
		}
		return instanceList
	}

	// find endpoint in subset by port
	for _, subset := range ep.Subsets {
		for _, endpointPort := range subset.Ports {
			if port.IntVal == endpointPort.Port || port.StrVal == endpointPort.Name {
				// add to subCluster
				for _, addr := range subset.Addresses {
					instanceList = append(instanceList, newBackendConf(addr.IP, int(endpointPort.Port), defaultWeight))
				}
			}
		}
	}

	return instanceList
}

// getTargetPort returns real targetport of backend pod
func getTargetPort(backendPort netv1.ServiceBackendPort, svc *corev1.Service) intstr.IntOrString {
	if svc == nil {
		log.Log.V(0).Info("service not found in getting target port")
		return intstr.IntOrString{}
	}

	// find matched port in service
	for _, p := range svc.Spec.Ports {
		if (backendPort.Number > 0 && backendPort.Number != p.Port) ||
			(len(backendPort.Name) > 0 && backendPort.Name != p.Name) {
			continue
		}

		// if name is present
		if len(p.Name) > 0 {
			return intstr.IntOrString{Type: intstr.String, StrVal: p.Name}
		}

		// if targetPort is present and in int format
		if p.TargetPort.IntVal > 0 {
			return intstr.IntOrString{Type: intstr.Int, IntVal: p.TargetPort.IntVal}
		}

		// targetPort is preset but in string format
		return intstr.IntOrString{}
	}

	return intstr.IntOrString{}
}

func newBackendConf(ip string, port int, weight int) *cluster_table_conf.BackendConf {
	return &cluster_table_conf.BackendConf{
		Name:   &ip,
		Addr:   &ip,
		Port:   &port,
		Weight: &weight,
	}
}

// makeGslbClusterConf makes cluster_table_conf.ClusterBackend configuration
func (c *ClusterConfig) newGslbClusterConf(namespace, service string, balance annotations.Balance) gslb_conf.GslbClusterConf {
	gslbConf := make(gslb_conf.GslbClusterConf)

	weights, ok := balance[service]
	if !ok {
		gslbConf[util.NamespacedName(namespace, service)] = defaultWeight
		return gslbConf
	}

	for name, weight := range weights {
		gslbConf[util.NamespacedName(namespace, name)] = weight
	}
	return gslbConf
}

func (c *ClusterConfig) UpdateService(service *corev1.Service, endpoint *corev1.Endpoints) error {
	serviceName := util.NamespacedName(service.Namespace, service.Name)

	// find cluster by service, do nothing if not found
	clusters, ok := c.service2Cluster.Get(serviceName)
	if !ok {
		return nil
	}

	// update cluster
	for _, cluster := range clusters {
		name := cluster.(string)

		targetPort := getTargetPort(util.ParsePort(name), service)
		// targetPort not found, which is not allowed for normal backend beside default backend
		if targetPort.IntVal == 0 && len(targetPort.StrVal) == 0 && serviceName != option.Opts.Ingress.DefaultBackend {
			c.DeleteService(service.Namespace, service.Name)
			log.Log.V(0).Info("ingress backend port not found in service", "namespace", service.Namespace, "name", service.Name, "port", util.ParsePort(name))
			return fmt.Errorf("cluster [%s] error, port can not found in service", name)
		} else {
			(*c.clusterTableConf.Config)[name][serviceName] = c.newSubClusterBackend(endpoint, targetPort)
			(*c.gslbConf.Clusters)[name] = c.newGslbClusterConf(service.Namespace, service.Name, nil)
		}
	}

	c.setVersion()
	return nil
}

func (c *ClusterConfig) DeleteService(namespace, name string) {
	serviceName := util.NamespacedName(namespace, name)

	// find cluster by service
	clusters, _ := c.service2Cluster.Get(serviceName)

	for _, cluster := range clusters {
		name := cluster.(string)

		// delete subcluster
		delete((*c.clusterTableConf.Config)[name], serviceName)

		// delete subcluter and cluster, total weigth of cluster should > 0
		delete((*c.gslbConf.Clusters)[name], serviceName)
		if len((*c.gslbConf.Clusters)[name]) == 0 {
			delete(*c.gslbConf.Clusters, name)
		}
	}

	c.setVersion()
}

func (c *ClusterConfig) Reload() error {
	reload := false
	if *c.gslbConf.Ts != c.gslbVersion {
		err := util.DumpBfeConf(GslbData, c.gslbConf)
		if err != nil {
			return fmt.Errorf("dump gslb.data error: %v", err)
		}

		reload = true
	}
	if *c.clusterTableConf.Version != c.clusterTableVersion {
		err := util.DumpBfeConf(ClusterTableData, c.clusterTableConf)
		if err != nil {
			return fmt.Errorf("dump cluster_table.data error: %v", err)
		}
		reload = true
	}

	if reload {
		if err := util.ReloadBfe(ConfigNameclusterConf); err != nil {
			return err
		}
		c.gslbVersion = *c.gslbConf.Ts
		c.clusterTableVersion = *c.clusterTableConf.Version
	}

	return nil
}
