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
	"io/ioutil"
	"os"
	"path"
)

import (
	"github.com/bfenetworks/bfe/bfe_util"
)

const (
	FilePerm os.FileMode = 0744
)

type Dumper struct {
	dumpRoot string // root path for dump
}

func NewDumper(root string) *Dumper {
	return &Dumper{
		dumpRoot: root,
	}
}

// DumpJson dumps json object to a relative path from dump root
func (d *Dumper) DumpJson(jsonObject interface{}, relativePath string) error {
	absolutePath := path.Join(d.dumpRoot, relativePath)
	return bfe_util.DumpJson(jsonObject, absolutePath, FilePerm)
}

// DumpBytes dumps byte data to a relative path from dump root
func (d *Dumper) DumpBytes(data []byte, relativePath string) error {
	absolutePath := path.Join(d.dumpRoot, relativePath)
	return ioutil.WriteFile(absolutePath, data, FilePerm)
}

// Join joins dump root and any number of suffix path elements into a single path,
// separating them with slashes. Empty elements are ignored.
// The result is Cleaned.
func (d *Dumper) Join(suffixes ...string) string {
	elem := []string{d.dumpRoot}
	elem = append(elem, suffixes...)
	return path.Join(elem[:]...)
}
