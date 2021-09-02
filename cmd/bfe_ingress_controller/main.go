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
	help            = flag.Bool("h", false, "to show help")
	stdOut          = flag.Bool("s", false, "to show log in stdout")
	showVersion     = flag.Bool("v", false, "to show version of BFE ingress controller")
	showVerbose     = flag.Bool("V", false, "to show verbose information")
	debugLog        = flag.Bool("d", false, "to show debug log (otherwise >= info)")
	logPath         = flag.String("l", "./log", "dir path of log")
	bfeConfigRoot   = flag.String("c", utils.DefaultBfeConfigRoot, "root dir path of BFE config")
	reloadURLPrefix = flag.String("u", utils.DefaultReloadURLPrefix, "BFE reload URL prefix")
	syncPeriod      = flag.Int("p", int(utils.DefaultSyncPeriod/time.Second),
		"sync period (in second) for Ingress watcher")
	namespaceLabels = flag.String("f", "", "namespace label selector, split by ,")
	ingressClass    = flag.String("k", "", "listen ingress class name")

	namespaces utils.Namespaces
)

var version string
var commit string

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
	if *showVerbose {
		printIngressVersion(version)
		fmt.Printf("go version: %s\n", runtime.Version())
		fmt.Printf("git commit: %s\n", commit)
		return
	}
	if *showVersion {
		printIngressVersion(version)
		return
	}

	// check ingress parameters
	if err := checkParams(); err != nil {
		fmt.Printf("bfe_ingress_controller: check params error[%s]", err.Error())
		return
	}

	// init log
	if err := initLog(); err != nil {
		fmt.Printf("bfe_ingress_controller: err in log.Init():%v\n", err)
		log.Logger.Close()
		os.Exit(1)
	}

	labels := strings.Split(*namespaceLabels, ",")

	// create BFE Ingress controller
	bfeIngress := bfe_ingress.NewBfeIngress(namespaces, labels, *ingressClass)
	bfeIngress.ReloadURLPrefix = *reloadURLPrefix
	bfeIngress.BfeConfigRoot = *bfeConfigRoot
	bfeIngress.SyncPeriod = time.Duration(*syncPeriod) * time.Second

	// start BFE Ingress controller
	log.Logger.Info("bfe_ingress_controller[version:%s] start", version)
	bfeIngress.Start()

	time.Sleep(1 * time.Second)
	log.Logger.Close()
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

func checkParams() error {
	if err := checkLabels(namespaces, *namespaceLabels); err != nil {
		return err
	}

	if *bfeConfigRoot == "" {
		return fmt.Errorf("BFE config root path should not be empty")
	}

	if *reloadURLPrefix == "" {
		return fmt.Errorf("BFE reload URL prefix should not be empty")
	}

	if *syncPeriod <= 0 {
		return fmt.Errorf("sync period for Ingress watcher sholud be greater then 0, period[%d]", *syncPeriod)
	}

	return nil
}

func printIngressVersion(version string) {
	fmt.Printf("bfe_ingress_controller version: %s\n", version)
}
