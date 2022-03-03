/*
Copyright 2020 The Kubernetes Authors.

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

package files

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Read tries to retrieve the desired file content from
// one of the registered file sources.
func Read(path string) ([]byte, error) {
	if exists := Exists(path); !exists {
		return nil, fmt.Errorf("file %v does not exists", path)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("fatal error retrieving test file %s: %s", path, err)
	}

	return data, nil
}

// Exists checks whether a file could be read. Unexpected errors
// are handled by calling the fail function, which then should
// abort the current test.
func Exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

// IsDir reports whether path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}
