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

package hostrules

import (
	"net/url"
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
	ctx.Step(`^a self-signed TLS secret named "([^"]*)" for the "([^"]*)" hostname$`, aSelfsignedTLSSecretNamedForTheHostname)
	ctx.Step(`^an Ingress resource$`, anIngressResource)
	ctx.Step(`^The Ingress status shows the IP address or FQDN where it is exposed$`, theIngressStatusShowsTheIPAddressOrFQDNWhereItIsExposed)
	ctx.Step(`^I send a "([^"]*)" request to "([^"]*)"$`, iSendARequestTo)
	ctx.Step(`^the secure connection must verify the "([^"]*)" hostname$`, theSecureConnectionMustVerifyTheHostname)
	ctx.Step(`^the response status-code must be (\d+)$`, theResponseStatuscodeMustBe)
	ctx.Step(`^the response must be served by the "([^"]*)" service$`, theResponseMustBeServedByTheService)
	ctx.Step(`^the request host must be "([^"]*)"$`, theRequestHostMustBe)

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

func aSelfsignedTLSSecretNamedForTheHostname(secretName string, host string) error {
	err := kubernetes.NewSelfSignedSecret(kubernetes.KubeClient, state.Namespace, secretName, []string{host})
	if err != nil {
		return err
	}

	state.SecretName = secretName

	return nil

}

func anIngressResource(spec *godog.DocString) error {
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

func iSendARequestTo(method string, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	return state.CaptureRoundTrip(method, u.Scheme, u.Host, u.Path, nil)
}

func theSecureConnectionMustVerifyTheHostname(hostname string) error {
	err := state.AssertTLSHostname(hostname)
	if err != nil {
		return err
	}

	err = state.AssertResponseCertificate(hostname)
	if err != nil {
		return err
	}

	return nil
}

func theResponseStatuscodeMustBe(statusCode int) error {
	return state.AssertStatusCode(statusCode)
}

func theResponseMustBeServedByTheService(service string) error {
	return state.AssertServedBy(service)
}

func theRequestHostMustBe(host string) error {
	return state.AssertRequestHost(host)
}
