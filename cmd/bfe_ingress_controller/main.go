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
	"runtime"
	"strings"
	"time"
)

import (
	"github.com/baidu/go-lib/log"
	"github.com/baidu/go-lib/log/log4go"
	"github.com/bfenetworks/bfe/bfe_debug"
	"github.com/bfenetworks/bfe/bfe_util"
	"github.com/bfenetworks/ingress-bfe/internal/bfe_ingress"
)

type Namespaces []string

func (n *Namespaces) String() string {
	return ""
}

func (n *Namespaces) Set(v string) error {
	*n = append(*n, v)
	return nil
}

var (
	help            *bool   = flag.Bool("h", false, "to show help")
	stdOut          *bool   = flag.Bool("s", false, "to show log in stdout")
	showVersion     *bool   = flag.Bool("v", false, "to show version of bfe_ingress_controller")
	showVerbose     *bool   = flag.Bool("V", false, "to show verbose information about bfe_ingress_controller")
	debugLog        *bool   = flag.Bool("d", false, "to show debug log (otherwise >= info)")
	logPath         *string = flag.String("l", "./log", "dir path of log")
	reloadUrlPrefix *string = flag.String("u", "http://localhost:8421/reload/", "reload URL prefix")
	configPath      *string = flag.String("c", "/bfe/output/conf/", "config path")
	resyncPeriod    *int    = flag.Int("p", 20, "resync period")
	namespaces      Namespaces
	namespaceLabels *string = flag.String("f", "", "namespace label selector, split by ,")

	ingressClass *string = flag.String("k", "", "listen ingress class name")
)

var version string
var commit string

func setIngressParams() {
	bfe_ingress.ReloadUrlPrefix = *reloadUrlPrefix
	bfe_ingress.ConfigPath = *configPath
	bfe_ingress.ResyncPeriod = time.Duration(*resyncPeriod) * time.Second
}

func checkLabels(namespaces Namespaces, labels string) error {
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
	var err error
	var logSwitch string

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

	if *debugLog {
		logSwitch = "DEBUG"
		bfe_debug.DebugIsOpen = true
	} else {
		logSwitch = "INFO"
		bfe_debug.DebugIsOpen = false
	}

	log4go.SetLogBufferLength(10000)
	log4go.SetLogWithBlocking(false)
	log4go.SetLogFormat(log4go.FORMAT_DEFAULT_WITH_PID)
	log4go.SetSrcLineForBinLog(false)
	err = log.Init("bfe_ingress_controller", logSwitch, *logPath, *stdOut, "midnight", 7)
	if err != nil {
		fmt.Printf("bfe_ingress_controller: err in log.Init():%v\n", err)
		bfe_util.AbnormalExit()
	}

	log.Logger.Info("bfe_ingress_controller[version:%s] start", version)
	setIngressParams()
	labels := strings.Split(*namespaceLabels, ",")
	bfe_ingress.WatchIngress(namespaces, labels, *ingressClass)

	time.Sleep(1 * time.Second)
	log.Logger.Close()
}
