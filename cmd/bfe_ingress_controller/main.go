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
	"os"
	"runtime"
	"strings"
	"time"
)

import (
	"github.com/baidu/go-lib/log"
	"github.com/baidu/go-lib/log/log4go"
)

import (
	"github.com/bfenetworks/ingress-bfe/internal/bfe_ingress"
	"github.com/bfenetworks/ingress-bfe/internal/utils"
)

var (
	help            *bool   = flag.Bool("h", false, "to show help")
	stdOut          *bool   = flag.Bool("s", false, "to show log in stdout")
	showVersion     *bool   = flag.Bool("v", false, "to show version of bfe_ingress_controller")
	showVerbose     *bool   = flag.Bool("V", false, "to show verbose information about bfe_ingress_controller")
	debugLog        *bool   = flag.Bool("d", false, "to show debug log (otherwise >= info)")
	logPath         *string = flag.String("l", "./log", "dir path of log")
	reloadUrlPrefix *string = flag.String("u", utils.ReloadUrlPrefix, "reload URL prefix")
	configPath      *string = flag.String("c", utils.ConfigPath, "config path")
	resyncPeriod    *int    = flag.Int("p", 20, "resync period")
	namespaces      utils.Namespaces
	namespaceLabels *string = flag.String("f", "", "namespace label selector, split by ,")

	ingressClass *string = flag.String("k", "", "listen ingress class name")
)

var version string
var commit string

func initIngressParams() {
	utils.ReloadUrlPrefix = *reloadUrlPrefix
	utils.ConfigPath = *configPath
	utils.ReSyncPeriod = time.Duration(*resyncPeriod) * time.Second
}

func initLog() error {
	var logSwitch string
	if *debugLog {
		logSwitch = "DEBUG"
	} else {
		logSwitch = "INFO"
	}

	log4go.SetLogBufferLength(10000)
	log4go.SetLogWithBlocking(false)
	log4go.SetLogFormat(log4go.FORMAT_DEFAULT_WITH_PID)
	log4go.SetSrcLineForBinLog(false)
	return log.Init("bfe_ingress_controller", logSwitch, *logPath, *stdOut, "midnight", 7)
}

func checkLabels(namespaces utils.Namespaces, labels string) error {
	if labels == "" {
		return nil
	}
	//namespace and label is exclusionary
	if len(namespaces) > 0 && labels != "" {
		return fmt.Errorf("labels and namespace sholud exclude, namespace[%s], labels[%s]", namespaces, labels)
	}
	labelsArr := strings.Split(labels, ",")
	for _, label := range labelsArr {
		keyValue := strings.Split(label, "=")
		if len(keyValue) != 2 {
			return fmt.Errorf("labels should be key=value, curVal[%s]", label)
		}
	}
	return nil
}

func main() {
	flag.Var(&namespaces, "n", "namespace to watch")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	if *showVersion {
		fmt.Printf("bfe_ingress_controller version: %s\n", version)
		return
	}
	if *showVerbose {
		fmt.Printf("bfe_ingress_controller version: %s\n", version)
		fmt.Printf("go version: %s\n", runtime.Version())
		fmt.Printf("git commit: %s\n", commit)
		return
	}
	if err := checkLabels(namespaces, *namespaceLabels); err != nil {
		fmt.Printf("bfe_ingress_controller: check params error[%s]", err.Error())
		return
	}

	// init log
	if err := initLog(); err != nil {
		fmt.Printf("bfe_ingress_controller: err in log.Init():%v\n", err)
		log.Logger.Close()
		os.Exit(1)
	}

	// init ingress parameters
	initIngressParams()

	labels := strings.Split(*namespaceLabels, ",")

	// create BFE Ingress controller
	bfeIngress := bfe_ingress.NewBfeIngress(namespaces, labels, *ingressClass)

	// start BFE Ingress controller
	log.Logger.Info("bfe_ingress_controller[version:%s] start", version)
	bfeIngress.Start()

	time.Sleep(1 * time.Second)
	log.Logger.Close()
}
