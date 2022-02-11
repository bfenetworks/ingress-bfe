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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"

	// ensure auth plugins are loaded
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// IngressClassValue sets the value of the class of Ingresses
var IngressClassValue string

// K8sNodeAddr sets ip addr of a k8s node
var K8sNodeAddr string

// KubeClient Kubernetes API client
var KubeClient *kubernetes.Clientset

var (
	BfeAnnotationPrefix = "bfe.ingress.kubernetes.io/"
	StatusAnnotationKey = fmt.Sprintf("%s%s", BfeAnnotationPrefix, "bfe-ingress-status")
)

type StatusMsg struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// LoadClientset returns clientset for connecting to kubernetes clusters.
func LoadClientset() (*clientset.Clientset, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		// Attempt to use local KUBECONFIG
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
		// use the current context in kubeconfig
		var err error

		config, err = kubeconfig.ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	// TODO: add version information?
	config.UserAgent = fmt.Sprintf(
		"%s (%s/%s) ingress-conformance",
		filepath.Base(os.Args[0]),
		runtime.GOOS,
		runtime.GOARCH,
	)

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewNamespace creates a new namespace using ingress-conformance- as prefix.
func NewNamespace(c kubernetes.Interface) (string, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ingress-conformance-",
			Labels: map[string]string{
				"app.kubernetes.io/name": "ingress-conformance",
			},
		},
	}

	var err error

	err = displayYamlDefinition(ns)
	if err != nil {
		return "", fmt.Errorf("unable show yaml definition: %v", err)
	}

	ns, err = c.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to create namespace: %v", err)
	}

	return ns.Name, nil
}

// DeleteNamespace deletes a namespace and all the objects inside
func DeleteNamespace(c kubernetes.Interface, namespace string) error {
	grace := int64(0)
	pb := metav1.DeletePropagationForeground

	if err := c.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &pb,
	}); err != nil {
		return err
	}

	return waitForNamespaceDeleted(c, namespace)
}

// DeleteNamespace deletes a namespace and all the objects inside
func DeleteNamespaceNonBlock(c kubernetes.Interface, namespace string) error {
	grace := int64(0)
	pb := metav1.DeletePropagationBackground

	return c.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &pb,
	})
}

// WaitForIngressAddress waits for the Ingress to acquire an address.
func waitForNamespaceDeleted(c clientset.Interface, namespace string) error {
	err := wait.PollImmediate(ingressWaitInterval, WaitForIngressAddressTimeout, func() (bool, error) {
		_, err := c.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if apierrs.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})

	return err
}

// CleanupNamespaces removes namespaces created by conformance tests
func CleanupNamespaces(c kubernetes.Interface) error {
	namespaces, err := c.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=ingress-conformance",
	})

	if err != nil {
		return err
	}

	for _, namespace := range namespaces.Items {
		err := DeleteNamespaceNonBlock(c, namespace.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewIngress creates a new ingress
func NewIngress(c kubernetes.Interface, namespace string, ingress *networking.Ingress) error {
	err := displayYamlDefinition(ingress)
	if err != nil {
		return fmt.Errorf("unable show yaml definition: %v", err)
	}

	if _, err := c.NetworkingV1().Ingresses(namespace).Create(context.TODO(), ingress, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

// IngressFromSpec deserializes an Ingress definition using an IngressSpec
func IngressFromSpec(name, namespace, ingressSpec string) (*networking.Ingress, error) {
	if namespace == metav1.NamespaceNone || namespace == metav1.NamespaceDefault {
		return nil, fmt.Errorf("ingress definitions in the default namespace are not allowed (%v)", namespace)
	}

	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := yaml.Unmarshal([]byte(ingressSpec), &ingress.Spec); err != nil {
		return nil, fmt.Errorf("deserializing Ingress from spec: %w", err)
	}

	if ingress.Spec.IngressClassName == nil {
		ingress.Spec.IngressClassName = &IngressClassValue
	}

	return ingress, nil
}

// IngressFromManifest deserializes an Ingress definition using an Ingress
func IngressFromManifest(namespace, manifest string) (*networking.Ingress, error) {
	if namespace == metav1.NamespaceNone || namespace == metav1.NamespaceDefault {
		return nil, fmt.Errorf("Ingress definitions in the default namespace are not allowed (%v)", namespace)
	}

	ingress := &networking.Ingress{}
	if err := yaml.Unmarshal([]byte(manifest), &ingress); err != nil {
		return nil, fmt.Errorf("deserializing Ingress from manifest: %w", err)
	}

	ingress.SetNamespace(namespace)

	if ingress.Spec.IngressClassName == nil {
		ingress.Spec.IngressClassName = &IngressClassValue
	}

	return ingress, nil
}

// NewSelfSignedSecret creates a self signed SSL certificate and store it in a secret
func NewSelfSignedSecret(c clientset.Interface, namespace, secretName string, hosts []string) error {
	if len(hosts) == 0 {
		return fmt.Errorf("require a non-empty hosts for Subject Alternate Name values")
	}

	var serverKey, serverCert bytes.Buffer

	host := strings.Join(hosts, ",")

	if err := generateRSACert(host, &serverKey, &serverCert); err != nil {
		return err
	}

	data := map[string][]byte{
		corev1.TLSCertKey:       serverCert.Bytes(),
		corev1.TLSPrivateKeyKey: serverKey.Bytes(),
	}

	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Type: corev1.SecretTypeTLS,
		Data: data,
	}

	err := displayYamlDefinition(newSecret)
	if err != nil {
		return fmt.Errorf("unable show yaml definition: %v", err)
	}

	if _, err := c.CoreV1().Secrets(namespace).Create(context.TODO(), newSecret, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

const (
	// ingressWaitInterval time to wait between checks for a condition
	ingressWaitInterval = 3 * time.Second
)

var (
	// WaitForIngressAddressTimeout maximum wait time for valid ingress status value
	WaitForIngressAddressTimeout = 3 * time.Minute
	// WaitForEndpointsTimeout maximum wait time for ready endpoints
	WaitForEndpointsTimeout = 3 * time.Minute

	// EnableOutputYamlDefinitions display yaml definitions of Kubernetes objects before creation
	EnableOutputYamlDefinitions = false

	// IngressControllerNameSpace the namespace of the ingress controller
	IngressControllerNameSpace = "ingress-bfe"
	// IngressControllerServiceName the service name of the ingress controller
	IngressControllerServiceName = "bfe-controller-service"
)

// WaitForIngressAddress waits for the Ingress to acquire an address.
func WaitForIngressAddress(c clientset.Interface, namespace, name string) (map[string]string, error) {
	var address map[string]string
	err := wait.PollImmediate(ingressWaitInterval, WaitForIngressAddressTimeout, func() (bool, error) {
		ipOrNameList, err := getIngressAddressOfNodePort(c, namespace, name)
		if err != nil || len(ipOrNameList) == 0 {
			if isRetryableAPIError(err) {
				return false, nil
			}

			return false, err
		}
		address = ipOrNameList
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("waiting for ingress status update: %w", err)
	}

	return address, nil
}

// getIngressAddress returns the ips/hostnames associated with the Ingress.
func getIngressAddress(c clientset.Interface, ns, name string) ([]string, error) {
	ing, err := c.NetworkingV1().Ingresses(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var addresses []string

	for _, a := range ing.Status.LoadBalancer.Ingress {
		if a.IP != "" {
			addresses = append(addresses, a.IP)
		}

		if a.Hostname != "" {
			addresses = append(addresses, a.Hostname)
		}
	}

	return addresses, nil
}

// getBfeIngressAddressByNodePort returns the service port name and port associated with the Ingress.
func getIngressAddressOfNodePort(c clientset.Interface, ns, name string) (map[string]string, error) {
	ing, err := c.NetworkingV1().Ingresses(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var statusMsg StatusMsg

	addresses := make(map[string]string)

	if ing.Annotations != nil {
		if statusMessage, ok := ing.Annotations[StatusAnnotationKey]; ok && statusMessage != "" {
			statusMessage := ing.Annotations[StatusAnnotationKey]
			err := json.Unmarshal([]byte(statusMessage), &statusMsg)
			if err != nil {
				return nil, err
			}
			if statusMsg.Status == "success" {
				service, err := c.CoreV1().Services(IngressControllerNameSpace).Get(context.TODO(), IngressControllerServiceName, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				ports := service.Spec.Ports
				for i := 0; i < len(ports); i++ {
					addresses[ports[i].Name] = fmt.Sprintf("%s:%d", K8sNodeAddr, ports[i].NodePort)
				}
			} else if statusMsg.Status == "error" {
				return nil, fmt.Errorf(statusMsg.Message)
			}
		}
	}
	return addresses, nil
}

func getIngressAddressOfClusterIp(c clientset.Interface, ns, name string) (map[string]string, error) {
	ing, err := c.NetworkingV1().Ingresses(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var statusMsg StatusMsg

	addresses := make(map[string]string)

	if ing.Annotations != nil {
		if statusMessage, ok := ing.Annotations[StatusAnnotationKey]; ok && statusMessage != "" {
			err := json.Unmarshal([]byte(statusMessage), &statusMsg)
			if err != nil {
				return nil, err
			}
			if statusMsg.Status == "success" {
				service, err := c.CoreV1().Services(IngressControllerNameSpace).Get(context.TODO(), "bfe-controller-service", metav1.GetOptions{})
				if err != nil {
					fmt.Println("error: ", err.Error())
					return nil, err
				}
				ports := service.Spec.Ports
				for i := 0; i < len(ports); i++ {
					addresses[ports[i].Name] = fmt.Sprintf("%s:%d", service.Spec.ClusterIP, service.Spec.Ports[i].Port)
				}
			} else {
				fmt.Printf("statusMessage: %s", statusMessage)
				return nil, fmt.Errorf("ingress status is not success")
			}
		}
	}
	fmt.Println("address", addresses)
	return addresses, nil
}

// isRetryableAPIError checks if an API error allows retries or not
func isRetryableAPIError(err error) bool {
	// These errors may indicate a transient error that we can retry in tests.
	if apierrs.IsInternalError(err) || apierrs.IsTimeout(err) || apierrs.IsServerTimeout(err) ||
		apierrs.IsTooManyRequests(err) || utilnet.IsProbableEOF(err) || utilnet.IsConnectionReset(err) {
		return true
	}

	// If the error sends the Retry-After header, we respect it as an explicit confirmation we should retry.
	if _, shouldRetry := apierrs.SuggestsClientDelay(err); shouldRetry {
		return true
	}

	// in case backend start slowly
	if err != nil && strings.Contains(err.Error(), "no avail backend") {
		return true
	}

	return false
}

const (
	rsaBits  = 2048
	validFor = 365 * 24 * time.Hour
)

// generateRSACert generates a basic self signed certificate valir for a year
func generateRSACert(host string, keyOut, certOut io.Writer) error {
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "default",
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	if err != nil {
		return fmt.Errorf("failed to create certificate: %s", err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return fmt.Errorf("failed creating key: %v", err)
	}

	return nil
}

func displayYamlDefinition(obj apiruntime.Object) error {
	if !EnableOutputYamlDefinitions {
		return nil
	}

	output, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, fmt.Sprintf("---\n%s\n", output))
	return err
}
