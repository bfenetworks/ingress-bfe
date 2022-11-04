// Copyright (c) 2022 The BFE Authors.
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

package redirect

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	tstate "github.com/bfenetworks/ingress-bfe/test/e2e/pkg/state"
	"github.com/cucumber/godog"

	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/kubernetes"
)

var state *tstate.Scenario

// IMPORTANT: Steps definitions are generated and should not be modified
// by hand but rather through make codegen. DO NOT EDIT.

// InitializeScenario configures the Feature to test
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^an Ingress resource with redirection annotations$`, anIngressResourceWithRedirectionAnnotations)
	ctx.Step(`^update the ingress by removing the redirect annotations$`, updateIngress)
	ctx.Step(`^The Ingress status shows the IP address or FQDN where it is exposed$`, theIngressStatusShowsTheIPAddressOrFQDNWhereItIsExposed)
	ctx.Step(`^I send a "([^"]*)" request to "([^"]*)"$`, iSendARequestTo)
	ctx.Step(`^the response status-code must be (\d+)$`, theResponseStatusCodeMustBe)
	ctx.Step(`^the response location must be "([^"]*)"$`, theResponseLocationMustBe)
	ctx.Step(`^The Ingress status should not be success$`, theIngressStatusShouldNotBeSuccess)

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		state = tstate.New()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		// delete namespace and all the content
		_ = kubernetes.DeleteNamespace(kubernetes.KubeClient, state.Namespace)
		return ctx, nil
	})
}

func anIngressResourceWithRedirectionAnnotations(spec *godog.DocString) error {
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

func updateIngress(spec *godog.DocString) error {
	ingress, err := kubernetes.IngressFromManifest(state.Namespace, spec.Content)
	if err != nil {
		return err
	}
	if ingress.GetName() != state.IngressName {
		return errors.New("ingress name is not match with the ingressName stored in the state when try to update")
	}

	err = kubernetes.UpdateIngress(kubernetes.KubeClient, state.Namespace, ingress)
	if err != nil {
		return err
	}

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

func theResponseLocationMustBe(location string) error {
	return state.AssertResponseHeader("Location", location)
}

func theIngressStatusShouldNotBeSuccess() error {
	_, err := kubernetes.WaitForIngressAddress(kubernetes.KubeClient, state.Namespace, state.IngressName)
	if err == nil {
		return fmt.Errorf("create ingress should return error")
	}

	return nil
}
