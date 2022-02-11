# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build
FROM golang:1.16-alpine3.14 as builder

WORKDIR /go/src/github.com/bfenetworks/ingress-bfe
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . /go/src/github.com/bfenetworks/ingress-bfe

# Build
RUN make e2e_test

FROM alpine3.14

ENV RESULTS_DIR="/tmp/results"
ENV WAIT_FOR_STATUS_TIMEOUT="5m"
ENV TEST_TIMEOUT="5m"

COPY --from=builder /go/src/github.com/bfenetworks/ingress-bfe/e2e_test /

COPY features /features
COPY run.sh /

CMD [ "/run.sh" ]
