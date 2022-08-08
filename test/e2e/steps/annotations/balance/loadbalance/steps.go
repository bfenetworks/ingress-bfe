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

package loadbalance

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/kubernetes"
	tstate "github.com/bfenetworks/ingress-bfe/test/e2e/pkg/state"
)

var (
	state *tstate.Scenario

	resultStatus  map[int]sets.String
	resultService sets.String
)

// IMPORTANT: Steps definitions are generated and should not be modified
// by hand but rather through make codegen. DO NOT EDIT.

// InitializeScenario configures the Feature to test
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^an Ingress with service info "([^"]*)" resource in a new random namespace$`, anIngressWithServiceInfoResourceInANewRandomNamespace)
	ctx.Step(`^The Ingress status shows the IP address or FQDN where it is exposed$`, theIngressStatusShowsTheIPAddressOrFQDNWhereItIsExposed)
	ctx.Step(`^I send (\d+) "([^"]*)" requests to "([^"]*)"$`, iSendRequestsTo)
	ctx.Step(`^the response status-code must be (\d+) the response body should contain the IP address of (\d+) different Kubernetes pods$`, theResponseStatuscodeMustBeTheResponseBodyShouldContainTheIPAddressOfDifferentKubernetesPods)
	ctx.Step(`^the response must be served by one of "([^"]*)" service$`, theResponseMustBeServedByOneOfService)
	ctx.Step(`^The Ingress status should not be success$`, theIngressStatusShouldNotBeSuccess)

	ctx.BeforeScenario(func(*godog.Scenario) {
		state = tstate.New()
		resultStatus = make(map[int]sets.String, 0)
		resultService = make(sets.String)
	})

	ctx.AfterScenario(func(*messages.Pickle, error) {
		// delete namespace an all the content
		_ = kubernetes.DeleteNamespace(kubernetes.KubeClient, state.Namespace)
	})
}

func anIngressWithServiceInfoResourceInANewRandomNamespace(serviceInfo string, spec *godog.DocString) error {
	ns, err := kubernetes.NewNamespace(kubernetes.KubeClient)
	if err != nil {
		return err
	}

	state.Namespace = ns

	ingress, err := kubernetes.IngressFromManifest(state.Namespace, spec.Content)
	if err != nil {
		return err
	}

	err = kubernetes.DeploymentsFromIngressForBalance(kubernetes.KubeClient, ingress, serviceInfo)
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

func iSendRequestsTo(totalRequest int, method string, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	for iteration := 1; iteration <= totalRequest; iteration++ {
		err := state.CaptureRoundTrip(method, u.Scheme, u.Host, u.Path, nil, nil, true)
		if err != nil {
			return err
		}

		if resultStatus[state.CapturedResponse.StatusCode] == nil {
			resultStatus[state.CapturedResponse.StatusCode] = sets.NewString()
		}

		resultStatus[state.CapturedResponse.StatusCode].Insert(state.CapturedRequest.Pod)
		resultService.Insert(state.CapturedRequest.Service)
	}

	return nil
}

func theResponseStatuscodeMustBeTheResponseBodyShouldContainTheIPAddressOfDifferentKubernetesPods(statusCode int, pods int) error {
	results, ok := resultStatus[statusCode]
	if !ok {
		return fmt.Errorf("no reponses for status code %v returned", statusCode)
	}

	if results.Len() != pods {
		return fmt.Errorf("expected %v different POD IP addresses/FQDN for status code %v but %v was returned", pods, statusCode, results.Len())
	}

	return nil
}

func theResponseMustBeServedByOneOfService(serviceInfo string) error {
	serviceList := strings.Split(serviceInfo, "|")
	for i := 0; i < len(serviceList); i++ {
		if !resultService.Has(serviceList[i]) {
			return fmt.Errorf("service info %s not exist in request info", serviceList[i])
		}
	}
	return nil
}

func theIngressStatusShouldNotBeSuccess() error {
	_, err := kubernetes.WaitForIngressAddress(kubernetes.KubeClient, state.Namespace, state.IngressName)
	if err == nil {
		return fmt.Errorf("create ingress should return error")
	}

	return nil
}
