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

package bfe_ingress

import (
	"io"
	"io/ioutil"
	"net/http"
)

import (
	"github.com/baidu/go-lib/log"
)

const (
	BodyLimit = 1024
)

var (
	ReloadUrlPrefix = "http://localhost:8421/reload/"
)

func SetReloadUrlPrefix(prefix string) {
	ReloadUrlPrefix = prefix
}

func reloadBfe(configName string) error {
	url := ReloadUrlPrefix + configName
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	failReason, err := ioutil.ReadAll(io.LimitReader(res.Body, BodyLimit))
	if err != nil {
		return err
	}

	log.Logger.Warn("Failed to reload %s: %s", configName, failReason)
	return nil
}
