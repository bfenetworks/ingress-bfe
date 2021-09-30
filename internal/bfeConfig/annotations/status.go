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
package annotations

import (
	"encoding/json"
)

var (
	StatusKey           = "bfe-ingress-status"
	StatusAnnotationKey = BfeAnnotationPrefix + StatusKey
)

type statusMsg struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func GenErrorMsg(err error) string {
	var status = statusMsg{}

	if err == nil {
		status.Status = "success"
	} else {
		status.Status = "error"
		status.Message = err.Error()
	}

	jsons, _ := json.Marshal(status)
	return string(jsons)
}

// CompareStatus check errMsg with status, return 0 if equal
func CompareStatus(e error, status string) int {
	if len(status) == 0 {
		return 1
	}

	s := &statusMsg{}
	if err := json.Unmarshal([]byte(status), s); err != nil {
		return 1
	}

	if e == nil && s.Status == "success" {
		return 0
	}

	if e != nil && s.Status == "error" && s.Message == e.Error() {
		return 0
	}

	return 1
}
