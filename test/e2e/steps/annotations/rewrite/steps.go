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

//Package rewrite defines the testing scenario and steps of rewrite url.
package rewrite

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"net/url"
	"time"

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
	ctx.Step(`^an Ingress resource with rewrite annotation$`, anIngressResourceWithRewriteAnnotation)
	ctx.Step(`^The Ingress status shows the IP address or FQDN where it is exposed$`, theIngressStatusShowsTheIPAddressOrFQDNWhereItIsExposed)
	ctx.Step(`^I send a "([^"]*)" request to "([^"]*)"$`, iSendARequestTo)
	ctx.Step(`^the response status code must be (\d+)$`, theResponseStatusCodeMustBe)
	ctx.Step(`^the request host must be "([^"]*)"$`, theRequestHostMustBe)
	ctx.Step(`^the request path must be "([^"]*)"$`, theRequestPathMustBe)
	ctx.Step(`^The Ingress status should not be success$`, theIngressStatusShouldNotBeSuccess)

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		state = tstate.New()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {

		_ = kubernetes.DeleteNamespace(kubernetes.KubeClient, state.Namespace)
		return ctx, nil
	})
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

func anIngressResourceWithRewriteAnnotation(spec *godog.DocString) error {
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

func iSendARequestTo(method string, rawURL string) error {
	u, err := url.Parse(rawURL)

	if err != nil {
		return err
	}
	return state.CaptureRoundTrip(method, "http", u.Host, u.Path, u.Query(), nil, false)
}

func theResponseStatusCodeMustBe(statusCode int) error {
	return state.AssertStatusCode(statusCode)
}

func theRequestHostMustBe(host string) error {
	return state.AssertRequestHost(host)
}

func theRequestPathMustBe(path string) error {
	return state.AssertRequestPath(path)
}

func theIngressStatusShouldNotBeSuccess() error {
	_, err := kubernetes.WaitForIngressAddress(kubernetes.KubeClient, state.Namespace, state.IngressName)
	if err == nil {
		return fmt.Errorf("create ingress should return error")
	}

	return nil
}
