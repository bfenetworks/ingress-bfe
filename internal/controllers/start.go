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

package controllers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/ingress"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/ingress/extv1beta1"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/ingress/netv1"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/ingress/netv1beta1"
	"github.com/bfenetworks/ingress-bfe/internal/option"
)

var (
	log = ctrl.Log.WithName("controllers")
)

func Start(scheme *runtime.Scheme) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("unable to get client config: %s", err)
	}

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     option.Opts.MetricsAddr,
		HealthProbeBindAddress: option.Opts.HealthProbeAddr,
	})
	if err != nil {
		return fmt.Errorf("unable to start controller manager: %s", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %s", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %s", err)
	}

	ctx := ctrl.SetupSignalHandler()

	// new bfe config builder
	cb := bfeConfig.NewConfigBuilder()
	cb.InitReload(ctx)

	// add controller to watch ingress resource
	if err := addController(cb, mgr); err != nil {
		return err
	}

	// start bfe process
	if err := startBFE(ctx); err != nil {
		return err
	}

	log.Info("starting manager")

	// start controller manager and blocking
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("fail to run manager: %s", err)
	}

	log.Info("exit manager")

	return nil
}

func addController(cb *bfeConfig.ConfigBuilder, mgr manager.Manager) error {
	client := discovery.NewDiscoveryClientForConfigOrDie(ctrl.GetConfigOrDie())
	serverVersion, err := client.ServerVersion()
	if err != nil {
		return fmt.Errorf("unable to get k8s cluster version: %s", err)
	}

	if serverVersion.Major >= "1" && serverVersion.Minor >= "19" {
		if err = netv1.AddIngressController(mgr, cb); err != nil {
			return fmt.Errorf("unable to create controller Ingress(netwokingv1): %s", err)
		}
	} else if serverVersion.Major >= "1" && serverVersion.Minor >= "14" {
		if err = netv1beta1.AddIngressController(mgr, cb); err != nil {
			return fmt.Errorf("unable to create controller Ingress(netwokingv1beta1): %s", err)
		}
	} else {
		if err = extv1beta1.AddIngressController(mgr, cb); err != nil {
			return fmt.Errorf("unable to create controller Ingress(extensionsv1beta1): %s", err)
		}
	}

	if err := ingress.AddServiceController(mgr, cb); err != nil {
		return fmt.Errorf("unable to create controller Service: %s", err)
	}

	if err := ingress.AddSecretController(mgr, cb); err != nil {
		return fmt.Errorf("unable to create controller secret: %s", err)
	}

	return nil
}

func startBFE(ctx context.Context) error {
	cmd := exec.Command(option.Opts.Ingress.BfeBinary, "-c", "../conf", "-l", "../log", "-s")
	cmd.Dir = filepath.Dir(option.Opts.Ingress.BfeBinary)

	log.Info("bfe is starting")

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("fail to start bfe: %s", err.Error())
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Error(err, "bfe exit")
		} else {
			log.Info("bfe exit")
		}

		// bfe exit, signaling controller to exit
		raise(syscall.SIGTERM)
	}()

	return err
}

func raise(sig os.Signal) error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(sig)
}
