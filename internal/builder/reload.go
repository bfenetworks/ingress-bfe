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

package builder

import (
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
)

import (
	"github.com/baidu/go-lib/log"
)

const (
	BodyLimit = 1024
)

type ConfigCache struct {
	lastConfigMap map[string]interface{}
	lock          sync.Mutex
}

func (c *ConfigCache) isConfigUpdated(key string, val interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	lastConfig, ok := c.lastConfigMap[key]
	if !ok {
		return true
	}
	return !reflect.DeepEqual(lastConfig, val)
}

func (c *ConfigCache) setConfig(key string, val interface{}) {
	c.lock.Lock()
	c.lastConfigMap[key] = val
	c.lock.Unlock()
}

type Reloader struct {
	urlPrefix string
	cache     *ConfigCache
}

func NewReloader(prefix string) *Reloader {
	return &Reloader{
		urlPrefix: prefix,
		cache: &ConfigCache{
			lastConfigMap: make(map[string]interface{}),
			lock:          sync.Mutex{},
		},
	}
}

func (r *Reloader) DoReload(newConfig interface{}, configName string) error {
	if !r.cache.isConfigUpdated(configName, newConfig) {
		return nil
	}
	log.Logger.Info("config[%s] has diff should reload", configName)
	r.cache.setConfig(configName, newConfig)
	r.reloadBfe(configName)
	return nil
}

func (r *Reloader) reloadBfe(configName string) error {
	url := r.urlPrefix + configName
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
		log.Logger.Warn("Failed to reload %s: %s", configName, failReason)
		return err
	}

	return nil
}
