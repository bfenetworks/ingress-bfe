# Copyright 2019 The Kubernetes Authors.
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

ENV CGO_ENABLED=0

WORKDIR /echoserver/

COPY echoserver.go .

RUN GO111MODULE=off go build -trimpath -ldflags="-buildid= -s -w" -o echoserver .

# Use distroless as minimal base image to package the binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.14
WORKDIR /
COPY --from=builder /echoserver /

ENTRYPOINT ["/echoserver"]
