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

package main

import (
	"flag"
	"fmt"
	rt "runtime"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig"
	"github.com/bfenetworks/ingress-bfe/internal/controllers"
	"github.com/bfenetworks/ingress-bfe/internal/option"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	initFlags()
}

var (
	version string
	commit  string
)

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if help {
		flag.PrintDefaults()
		return
	}
	if showVersion {
		fmt.Printf("bfe-ingress-controller version: %s\n", version)
		fmt.Printf("go version: %s\n", rt.Version())
		fmt.Printf("git commit: %s\n", commit)
		return
	}

	err := option.SetOptions(
		namespaces, ingressClass, configPath, reloadAddr,
		metricsAddr, probeAddr, defaultBackend)
	if err != nil {
		setupLog.Error(err, "fail to start controllers")
		return
	}

	setupLog.Info("starting bfe-ingress-controller")

	configBuilder := bfeConfig.NewConfigBuilder()
	configBuilder.InitReload()

	if err := controllers.Start(scheme, configBuilder); err != nil {
		setupLog.Error(err, "fail to start controllers")
	}

	setupLog.Info("bfe-ingress-controller exit")
}
