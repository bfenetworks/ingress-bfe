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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/corev1"
	extensionsv1beta1 "github.com/bfenetworks/ingress-bfe/internal/controllers/extv1beta1"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/netv1"
	"github.com/bfenetworks/ingress-bfe/internal/controllers/netv1beta1"
	option "github.com/bfenetworks/ingress-bfe/internal/option"
)

var (
	log = ctrl.Log.WithName("controllers")
)

func Start(scheme *runtime.Scheme, configBuilder *bfeConfig.ConfigBuilder) error {
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

	if err = (&corev1.EndpointsReconciler{
		BfeConfigBuilder: configBuilder,
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller Endpoints: %s", err)
	}

	if err = (&corev1.SecretReconciler{
		BfeConfigBuilder: configBuilder,
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create controller secret: %s", err)
	}

	client := discovery.NewDiscoveryClientForConfigOrDie(ctrl.GetConfigOrDie())
	serverVersion, err := client.ServerVersion()
	if err != nil {
		return fmt.Errorf("unable to get k8s cluster version: %s", err)
	}

	if serverVersion.Major >= "1" && serverVersion.Minor >= "19" {
		if err = (&netv1.IngressReconciler{
			BfeConfigBuilder: configBuilder,
			Client:           mgr.GetClient(),
			Scheme:           mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create controller Ingress(netwokingv1): %s", err)
		}
	} else if serverVersion.Major >= "1" && serverVersion.Minor >= "14" {
		if err = (&netv1beta1.IngressReconciler{
			BfeConfigBuilder: configBuilder,
			Client:           mgr.GetClient(),
			Scheme:           mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create controller Ingress(netwokingv1beta1): %s", err)
		}
	} else {
		if err = (&extensionsv1beta1.IngressReconciler{
			BfeConfigBuilder: configBuilder,
			Client:           mgr.GetClient(),
			Scheme:           mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			return fmt.Errorf("unable to create controller Ingress(extensionsv1beta1): %s", err)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %s", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %s", err)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("fail to run manager: %s", err)
	}
	log.Info("existing manager")

	return nil
}
