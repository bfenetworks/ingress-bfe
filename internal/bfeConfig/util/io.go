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
package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bfenetworks/ingress-bfe/internal/option"
)

func DumpBfeConf(configFile string, object interface{}) error {
	buf, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return fmt.Errorf("config json marshal err %s", err)
	}
	return DumpFile(configFile, buf)
}

func DumpFile(filename string, data []byte) error {
	name := option.Opts.ConfigPath + filename
	filePath := filepath.Dir(name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.MkdirAll(filePath, option.FilePerm)
	}

	return ioutil.WriteFile(name, data, option.FilePerm)
}

func DeleteFile(filename string) {
	name := option.Opts.ConfigPath + filename
	os.Remove(name)
}

// ReloadBfe triggers bfe process to reload new config file through bfe monitor port
func ReloadBfe(configName string) error {
	url := option.Opts.ReloadUrl + configName
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	failReason, err := ioutil.ReadAll(io.LimitReader(res.Body, 1024))
	if err != nil {
		return err
	}

	return fmt.Errorf("fail to reload: %s", failReason)
}
