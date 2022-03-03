/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/kubernetes/templates"
)

// EchoService name of the deployment for the echo app
const EchoService = "echo"

// EchoContainer container image name
const EchoContainer = "local/echoserver:0.0.1"

// NewEchoDeployment creates a new deployment of the echoserver image in a particular namespace.
func NewEchoDeployment(kubeClientSet kubernetes.Interface, namespace, name, serviceName, servicePortName string, servicePort int32) error {
	deploymentName := fmt.Sprintf("%v-%v", name, serviceName)

	deployment, err := kubeClientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		// if the deployment doesn't exists is still returned
		deployment = nil
	}

	// assume an existing deployment is ok
	if deployment != nil {
		return nil
	}

	deploymentData := struct {
		Name        string
		MatchLabels string
		Labels      string
		Image       string
		Ingress     string
		Service     string
		PortName    string
	}{
		deploymentName,
		deploymentName,
		deploymentName,
		EchoContainer,
		name,
		serviceName,
		servicePortName,
	}

	manifest, err := templates.Render("deployment", deploymentData)
	if err != nil {
		return err
	}

	deployment, err = deploymentFromManifest(manifest)
	if err != nil {
		return err
	}

	err = displayYamlDefinition(deployment)
	if err != nil {
		return fmt.Errorf("unable show yaml definition: %v", err)
	}

	_, err = kubeClientSet.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating deployment (%v): %w", deployment.Name, err)
	}

	serviceData := struct {
		Name     string
		Selector string
		Port     int32
	}{
		serviceName,
		deploymentName,
		servicePort,
	}

	manifest, err = templates.Render("service", serviceData)
	if err != nil {
		return err
	}

	service, err := serviceFromManifest(manifest)
	if err != nil {
		return err
	}

	if servicePortName != "" {
		service.Spec.Ports[0].Name = servicePortName
	}

	// if no port is defined, use default 8080
	if servicePort == 0 {
		service.Spec.Ports[0].Port = 8080
	}

	err = displayYamlDefinition(service)
	if err != nil {
		return fmt.Errorf("unable show yaml definition: %v", err)
	}

	service, err = kubeClientSet.CoreV1().Services(namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating service (%v): %w", service.Name, err)
	}

	err = waitForEndpoints(kubeClientSet, WaitForEndpointsTimeout, service.Namespace, service.Name, 1)
	if err != nil {
		return fmt.Errorf("waiting for service (%v) endpoints available: %w", service.Name, err)
	}

	return nil
}

// DeploymentsFromIngress creates the required deployments for the services defined in the ingress object
func DeploymentsFromIngress(kubeClientSet kubernetes.Interface, ingress *networking.Ingress) error {

	if ingress.Spec.DefaultBackend != nil {
		serviceName := ingress.Spec.DefaultBackend.Service.Name
		servicePort := ingress.Spec.DefaultBackend.Service.Port.Number

		err := NewEchoDeployment(kubeClientSet, ingress.Namespace, ingress.Name, serviceName, "", servicePort)
		if err != nil {
			return err
		}
	}

	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		for _, path := range rule.HTTP.Paths {
			serviceName := path.Backend.Service.Name
			servicePort := path.Backend.Service.Port.Number

			err := NewEchoDeployment(kubeClientSet, ingress.Namespace, ingress.Name, serviceName, "", servicePort)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeploymentsFromIngressForBalance creates the required deployments for the services defined in the ingress object or in the param of service info
func DeploymentsFromIngressForBalance(kubeClientSet kubernetes.Interface, ingress *networking.Ingress, serviceInfo string) error {

	if ingress.Spec.DefaultBackend != nil {
		serviceName := ingress.Spec.DefaultBackend.Service.Name
		servicePort := ingress.Spec.DefaultBackend.Service.Port.Number

		err := NewEchoDeployment(kubeClientSet, ingress.Namespace, ingress.Name, serviceName, "", servicePort)
		if err != nil {
			return err
		}
	}

	if serviceInfo != "" {
		serviceNameList := strings.Split(serviceInfo, "|")
		for i := 0; i < len(serviceNameList); i++ {
			ipPort := strings.Split(serviceNameList[i], ":")
			if len(ipPort) < 2 {
				return fmt.Errorf("error ip port for service info")
			}
			serviceName := ipPort[0]
			servicePort, err := strconv.ParseInt(ipPort[1], 10, 32)
			if err != nil {
				return err
			}
			err = NewEchoDeployment(kubeClientSet, ingress.Namespace, ingress.Name, serviceName, "", int32(servicePort))
			if err != nil {
				return err
			}
		}
	} else {
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				serviceName := path.Backend.Service.Name
				servicePort := path.Backend.Service.Port.Number

				err := NewEchoDeployment(kubeClientSet, ingress.Namespace, ingress.Name, serviceName, "", servicePort)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ScaleIngressBackendDeployment changes the replicas count of a deployment defined in an ingress service backend
func ScaleIngressBackendDeployment(kubeClientSet kubernetes.Interface, namespace, name, serviceName string, replicas int) error {
	deploymentName := fmt.Sprintf("%v-%v", name, serviceName)

	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: int32(replicas),
		},
	}

	_, err := kubeClientSet.AppsV1().Deployments(namespace).UpdateScale(context.TODO(), deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	err = waitForEndpoints(kubeClientSet, WaitForEndpointsTimeout, namespace, serviceName, replicas)
	if err != nil {
		return fmt.Errorf("waiting for service (%v) endpoints available: %w", serviceName, err)
	}

	time.Sleep(60 * time.Second)

	return nil
}

// deploymentFromManifest deserializes a Deployment definition from a yaml string
func deploymentFromManifest(manifest string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if err := yaml.Unmarshal([]byte(manifest), &deployment); err != nil {
		return nil, fmt.Errorf("deserializing deployment from manifest: %w\n%v", err, manifest)
	}

	return deployment, nil
}

// serviceFromManifest deserializes a Deployment definition from a yaml string
func serviceFromManifest(manifest string) (*corev1.Service, error) {
	deployment := &corev1.Service{}
	if err := yaml.Unmarshal([]byte(manifest), &deployment); err != nil {
		return nil, fmt.Errorf("deserializing service from manifest: %w", err)
	}

	return deployment, nil
}

// waitForEndpoints waits for a given amount of time until the number of endpoints = expectedEndpoints.
func waitForEndpoints(kubeClientSet kubernetes.Interface, timeout time.Duration, ns, name string, expectedEndpoints int) error {
	if expectedEndpoints == 0 {
		return nil
	}

	return wait.Poll(5*time.Second, timeout, func() (bool, error) {
		endpoint, err := kubeClientSet.CoreV1().Endpoints(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		if countReadyEndpoints(endpoint) == expectedEndpoints {
			return true, nil
		}

		return false, nil
	})
}

func countReadyEndpoints(e *corev1.Endpoints) int {
	if e == nil || e.Subsets == nil {
		return 0
	}

	num := 0
	for _, sub := range e.Subsets {
		num += len(sub.Addresses)
	}

	return num
}
