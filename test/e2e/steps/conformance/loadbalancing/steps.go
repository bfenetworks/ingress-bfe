/*
Copyright 2021 The BFE Authors.

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

package loadbalancing

import (
	"fmt"
	"net/url"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/kubernetes"
	tstate "github.com/bfenetworks/ingress-bfe/test/e2e/pkg/state"
)

var (
	state *tstate.Scenario

	resultStatus map[int]sets.String
)

// IMPORTANT: Steps definitions are generated and should not be modified
// by hand but rather through make codegen. DO NOT EDIT.

// InitializeScenario configures the Feature to test
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^an Ingress resource in a new random namespace$`, anIngressResourceInANewRandomNamespace)
	ctx.Step(`^The Ingress status shows the IP address or FQDN where it is exposed$`, theIngressStatusShowsTheIPAddressOrFQDNWhereItIsExposed)
	ctx.Step(`^The backend deployment "([^"]*)" for the ingress resource is scaled to (\d+)$`, theBackendDeploymentForTheIngressResourceIsScaledTo)
	ctx.Step(`^I send (\d+) requests to "([^"]*)"$`, iSendRequestsTo)
	ctx.Step(`^all the responses status-code must be (\d+) and the response body should contain the IP address of (\d+) different Kubernetes pods$`, allTheResponsesStatuscodeMustBeAndTheResponseBodyShouldContainTheIPAddressOfDifferentKubernetesPods)

	ctx.BeforeScenario(func(*godog.Scenario) {
		state = tstate.New()
		resultStatus = make(map[int]sets.String, 0)
	})

	ctx.AfterScenario(func(*messages.Pickle, error) {
		// delete namespace an all the content
		_ = kubernetes.DeleteNamespace(kubernetes.KubeClient, state.Namespace)
	})
}

func anIngressResourceInANewRandomNamespace(spec *godog.DocString) error {
	ns, err := kubernetes.NewNamespace(kubernetes.KubeClient)
	if err != nil {
		return err
	}

	state.Namespace = ns

	ingress, err := kubernetes.IngressFromManifest(state.Namespace, spec.Content)
	if err != nil {
		return err
	}

	err = kubernetes.DeploymentsFromIngress(kubernetes.KubeClient, ingress)
	if err != nil {
		return err
	}

	err = kubernetes.NewIngress(kubernetes.KubeClient, state.Namespace, ingress)
	if err != nil {
		return err
	}

	state.IngressName = ingress.GetName()

	return nil
}

func theIngressStatusShowsTheIPAddressOrFQDNWhereItIsExposed() error {
	ingress, err := kubernetes.WaitForIngressAddress(kubernetes.KubeClient, state.Namespace, state.IngressName)
	if err != nil {
		return err
	}

	state.IPOrFQDN = ingress

	time.Sleep(3 * time.Second)

	return err
}

func theBackendDeploymentForTheIngressResourceIsScaledTo(deployment string, replicas int) error {
	return kubernetes.ScaleIngressBackendDeployment(kubernetes.KubeClient, state.Namespace, state.IngressName, deployment, replicas)
}

func iSendRequestsTo(totalRequest int, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	for iteration := 1; iteration <= totalRequest; iteration++ {
		err := state.CaptureRoundTrip("GET", u.Scheme, u.Host, u.Path, nil, nil, true)
		if err != nil {
			return err
		}

		if resultStatus[state.CapturedResponse.StatusCode] == nil {
			resultStatus[state.CapturedResponse.StatusCode] = sets.NewString()
		}

		resultStatus[state.CapturedResponse.StatusCode].Insert(state.CapturedRequest.Pod)
	}

	return nil
}

func allTheResponsesStatuscodeMustBeAndTheResponseBodyShouldContainTheIPAddressOfDifferentKubernetesPods(statusCode int, pods int) error {
	results, ok := resultStatus[statusCode]
	if !ok {
		return fmt.Errorf("no reponses for status code %v returned", statusCode)
	}

	if results.Len() != pods {
		return fmt.Errorf("expected %v different POD IP addresses/FQDN for status code %v but %v was returned", pods, statusCode, results.Len())
	}

	return nil
}
