/*
Copyright 2020 The BFE Authors.

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

package e2e

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/http"
	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/kubernetes"
	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/kubernetes/templates"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/annotations/balance/loadbalance"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/annotations/route/cookie"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/annotations/route/header"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/annotations/route/priority"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/conformance/hostrules"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/conformance/ingressclass"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/conformance/loadbalancing"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/conformance/pathrules"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/rules/host"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/rules/multipleingress"
	rules_path "github.com/bfenetworks/ingress-bfe/test/e2e/steps/rules/path"
	"github.com/bfenetworks/ingress-bfe/test/e2e/steps/rules/patherr"
)

var (
	godogFormat        string
	godogTags          string
	godogStopOnFailure bool
	godogNoColors      bool
	godogOutput        string
	godogTestFeature   string
)

func TestMain(m *testing.M) {
	// register flags from klog (client-go verbose logging)
	klog.InitFlags(nil)

	flag.StringVar(&godogFormat, "format", "pretty", "Set godog format to use. Valid values are pretty and cucumber")
	flag.StringVar(&godogTags, "tags", "", "Tags for e2e test")
	flag.BoolVar(&godogStopOnFailure, "stop-on-failure", false, "Stop when failure is found")
	flag.BoolVar(&godogNoColors, "no-colors", false, "Disable colors in godog output")
	flag.StringVar(&godogOutput, "output-directory", ".", "Output directory for test reports")
	flag.StringVar(&kubernetes.IngressClassValue, "ingress-class", "bfe", "Sets the value of the annotation kubernetes.io/ingress.class in Ingress definitions")
	flag.DurationVar(&kubernetes.WaitForIngressAddressTimeout, "wait-time-for-ingress-status", 3*time.Minute, "Maximum wait time for valid ingress status value")
	flag.DurationVar(&kubernetes.WaitForEndpointsTimeout, "wait-time-for-ready", 3*time.Minute, "Maximum wait time for ready endpoints")
	flag.StringVar(&kubernetes.IngressControllerNameSpace, "ingress-controller-namespace", "ingress-bfe", "Sets the value of the namespace for ingress controller")
	flag.StringVar(&kubernetes.IngressControllerServiceName, "ingress-controller-service-name", "bfe-controller-service", "Sets the name of the service for ingress controller")
	flag.StringVar(&kubernetes.K8sNodeAddr, "k8s-node-addr", "127.0.0.1", "Sets the ip address of one k8s node")
	flag.StringVar(&godogTestFeature, "feature", "", "Sets the file to test")
	flag.BoolVar(&http.EnableDebug, "enable-http-debug", false, "Enable dump of requests and responses of HTTP requests (useful for debug)")
	flag.BoolVar(&kubernetes.EnableOutputYamlDefinitions, "enable-output-yaml-definitions", false, "Dump yaml definitions of Kubernetes objects before creation")

	flag.Parse()

	validFormats := sets.NewString("cucumber", "pretty")
	if !validFormats.Has(godogFormat) {
		klog.Fatalf("the godog format '%v' is not supported", godogFormat)
	}

	err := setup()
	if err != nil {
		klog.Fatal(err)
	}

	if err := kubernetes.CleanupNamespaces(kubernetes.KubeClient); err != nil {
		klog.Fatalf("error deleting temporal namespaces: %v", err)
	}

	go handleSignals()

	os.Exit(m.Run())
}

func setup() error {
	err := templates.Load()
	if err != nil {
		return fmt.Errorf("error loading templates: %v", err)
	}

	kubernetes.KubeClient, err = kubernetes.LoadClientset()
	if err != nil {
		return fmt.Errorf("error loading client: %v", err)
	}

	return nil
}

var (
	features = map[string]func(*godog.ScenarioContext){
		"features/conformance/host_rules.feature":           hostrules.InitializeScenario,
		"features/conformance/ingress_class.feature":        ingressclass.InitializeScenario,
		"features/conformance/load_balancing.feature":       loadbalancing.InitializeScenario,
		"features/conformance/path_rules.feature":           pathrules.InitializeScenario,
		"features/rules/host_rule1.feature":                 host.InitializeScenario,
		"features/rules/host_rule2.feature":                 host.InitializeScenario,
		"features/rules/multiple_ingress.feature":           multipleingress.InitializeScenario,
		"features/rules/path_rule1.feature":                 rules_path.InitializeScenario,
		"features/rules/path_rule2.feature":                 rules_path.InitializeScenario,
		"features/rules/path_err.feature":                   patherr.InitializeScenario,
		"features/annotations/route/cookie.feature":         cookie.InitializeScenario,
		"features/annotations/route/header.feature":         header.InitializeScenario,
		"features/annotations/route/priority.feature":       priority.InitializeScenario,
		"features/annotations/balance/load_balance.feature": loadbalance.InitializeScenario,
	}
)

func TestSuite(t *testing.T) {
	var failed bool

	activeFeatures := make(map[string]func(*godog.ScenarioContext))
	for file, init := range features {
		if strings.HasPrefix(file, godogTestFeature) {
			activeFeatures[file] = init
		}
	}

	for feature, scenarioContext := range activeFeatures {
		err := testFeature(feature, scenarioContext)
		if err != nil {
			if godogStopOnFailure {
				t.Fatal(err)
			}
			failed = true
		}
	}

	if failed {
		t.Fatal("at least one step/scenario failed")
	}

}

func testFeature(feature string, scenarioInitializer func(*godog.ScenarioContext)) error {
	var testOutput io.Writer
	// default output is stdout
	testOutput = os.Stdout

	if godogFormat == "cucumber" {
		rf := path.Join(godogOutput, fmt.Sprintf("%v-report.json", filepath.Base(feature)))
		file, err := os.Create(rf)
		if err != nil {
			return fmt.Errorf("error creating report file %v: %w", rf, err)
		}

		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()

		testOutput = writer
	}

	opts := godog.Options{
		Format:        godogFormat,
		Paths:         []string{feature},
		Tags:          godogTags,
		StopOnFailure: godogStopOnFailure,
		NoColors:      godogNoColors,
		Output:        testOutput,
		Concurrency:   1, // do not run tests concurrently
	}

	exitCode := godog.TestSuite{
		Name:                "e2e-test",
		ScenarioInitializer: scenarioInitializer,
		Options:             &opts,
	}.Run()
	if exitCode > 0 {
		return fmt.Errorf("unexpected exit code testing %v: %v", feature, exitCode)
	}

	return nil
}

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	if err := kubernetes.CleanupNamespaces(kubernetes.KubeClient); err != nil {
		klog.Fatalf("error deleting temporal namespaces: %v", err)
	}

	os.Exit(1)
}
