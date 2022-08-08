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

package defaultbackend

import (
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"

	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/kubernetes"
	tstate "github.com/bfenetworks/ingress-bfe/test/e2e/pkg/state"
)

var (
	state *tstate.Scenario
)

// IMPORTANT: Steps definitions are generated and should not be modified
// by hand but rather through make codegen. DO NOT EDIT.

// InitializeScenario configures the Feature to test
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a new random namespace$`, aNewRandomNamespace)
	ctx.Step(`^an Ingress resource named "([^"]*)" with this spec:$`, anIngressResourceNamedWithThisSpec)
	ctx.Step(`^The Ingress status shows the IP address or FQDN where it is exposed$`, theIngressStatusShowsTheIPAddressOrFQDNWhereItIsExposed)
	ctx.Step(`^I send a "([^"]*)" request to http:\/\/"([^"]*)"\/"([^"]*)"$`, iSendARequestToHttp)
	ctx.Step(`^the response status-code must be (\d+)$`, theResponseStatuscodeMustBe)
	ctx.Step(`^the response must be served by the "([^"]*)" service$`, theResponseMustBeServedByTheService)
	ctx.Step(`^the response proto must be "([^"]*)"$`, theResponseProtoMustBe)
	ctx.Step(`^the response headers must contain <key> with matching <value>$`, theResponseHeadersMustContainKeyWithMatchingValue)
	ctx.Step(`^the request method must be "([^"]*)"$`, theRequestMethodMustBe)
	ctx.Step(`^the request path must be "([^"]*)"$`, theRequestPathMustBe)
	ctx.Step(`^the request proto must be "([^"]*)"$`, theRequestProtoMustBe)
	ctx.Step(`^the request headers must contain <key> with matching <value>$`, theRequestHeadersMustContainKeyWithMatchingValue)

	ctx.BeforeScenario(func(*godog.Scenario) {
		state = tstate.New()
	})

	ctx.AfterScenario(func(*messages.Pickle, error) {
		// delete namespace an all the content
		_ = kubernetes.DeleteNamespace(kubernetes.KubeClient, state.Namespace)
	})
}

func aNewRandomNamespace() error {
	ns, err := kubernetes.NewNamespace(kubernetes.KubeClient)
	if err != nil {
		return err
	}

	state.Namespace = ns
	return nil
}

func anIngressResourceNamedWithThisSpec(name string, spec *godog.DocString) error {
	ingress, err := kubernetes.IngressFromSpec(name, state.Namespace, spec.Content)
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

	state.IngressName = name

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

func iSendARequestToHttp(method string, hostname string, path string) error {
	return state.CaptureRoundTrip(method, "http", hostname, path, nil, nil, true)
}

func theResponseStatuscodeMustBe(statusCode int) error {
	return state.AssertStatusCode(statusCode)
}

func theResponseMustBeServedByTheService(service string) error {
	return state.AssertServedBy(service)
}

func theResponseProtoMustBe(proto string) error {
	return state.AssertResponseProto(proto)
}

func theResponseHeadersMustContainKeyWithMatchingValue(headers *godog.Table) error {
	return assertHeaderTable(headers, state.AssertResponseHeader)
}

func theRequestMethodMustBe(method string) error {
	return state.AssertMethod(method)
}

func theRequestPathMustBe(path string) error {
	return state.AssertRequestPath(path)
}

func theRequestProtoMustBe(proto string) error {
	return state.AssertRequestProto(proto)
}

func theRequestHeadersMustContainKeyWithMatchingValue(headers *godog.Table) error {
	return assertHeaderTable(headers, state.AssertRequestHeader)
}

func assertHeaderTable(headerTable *godog.Table, assertF func(key string, value string) error) error {
	if len(headerTable.Rows) < 1 {
		return fmt.Errorf("expected a table with at least one row")
	}

	for i, row := range headerTable.Rows {
		if len(row.Cells) != 2 {
			return fmt.Errorf("expected a table with 2 cells, it contained %v", len(row.Cells))
		}

		headerKey := row.Cells[0].Value
		headerValue := row.Cells[1].Value

		if i == 0 {
			if headerKey != "key" && headerValue != "value" {
				return fmt.Errorf("expected a table with a header row of 'key' and 'value' but got '%v' and '%v'", headerKey, headerValue)
			}
			// Skip the header row
			continue
		}

		if err := assertF(headerKey, headerValue); err != nil {
			return err
		}
	}

	return nil
}
