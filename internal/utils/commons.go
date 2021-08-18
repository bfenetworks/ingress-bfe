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

package utils

import (
	"os"
	"strings"
	"time"
)

const (
	FilePerm os.FileMode = 0744
)

// Namespaces implements flag.Value interface as []string
type Namespaces []string

// String implements flag.Value.String()
func (n *Namespaces) String() string {
	return strings.Join(*n, ",")
}

// Set implements flag.Value.Set()
func (n *Namespaces) Set(v string) error {
	*n = append(*n, v)
	return nil
}

var (
	ReloadUrlPrefix = "http://localhost:8421/reload/"
	ReSyncPeriod    = 20 * time.Second
	ConfigPath      = "/bfe/output/conf/"
)
