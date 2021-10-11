# Copyright (c) 2021 The BFE Authors.
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

# init project path
WORKROOT := $(shell pwd)
OUTDIR   := $(WORKROOT)/output

# init environment variables
export PATH        := $(shell go env GOPATH)/bin:$(PATH)
export GO111MODULE := on

# init command params
GO           := go
GOBUILD      := $(GO) build
GOTEST       := $(GO) test
GOVET        := $(GO) vet
GOGET        := $(GO) get
GOGEN        := $(GO) generate
GOCLEAN      := $(GO) clean
GOFLAGS      := -race
STATICCHECK  := staticcheck

# init arch
ARCH := $(shell getconf LONG_BIT)
ifeq ($(ARCH),64)
	GOTEST += $(GOFLAGS)
endif

# init bfe ingress version
INGRESS_VERSION ?= $(shell cat VERSION)
# init git commit id
GIT_COMMIT ?= $(shell git rev-parse HEAD)

# init bfe ingress
INGRESS_PACKAGES := $(shell go list ./...)

# make, make all
all: compile 

# make compile, go build
compile: test build
build:
	$(WORKROOT)/build/build.sh

# make test, test your code
test: test-case vet-case
test-case:
	$(GOTEST) -cover ./...
vet-case:
	${GOVET} ./...

# make coverage for codecov
coverage:
	echo -n > coverage.txt
	for pkg in $(INGRESS_PACKAGES) ; do $(GOTEST) -coverprofile=profile.out -covermode=atomic $${pkg} && cat profile.out >> coverage.txt; done

# make check
check:
	$(GO) get honnef.co/go/tools/cmd/staticcheck
	$(STATICCHECK) ./...

# make docker
docker:
	docker build \
		-t bfenetworks/bfe-ingress-controller:$(INGRESS_VERSION) \
		-f Dockerfile \
		.

# make clean
clean:
	$(GOCLEAN)
	rm -rf $(OUTDIR)

# avoid filename conflict and speed up build
.PHONY: all compile test clean build