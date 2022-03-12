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

.DEFAULT_GOAL:=help

MKDIR_P := mkdir -p
RM_F := rm -rf


PROGRAMS := \
	e2e_test

TAG ?= 0.0.1

REGISTRY ?= local

build: $(PROGRAMS) ## Build the e2e-test binary

.PHONY: build-image
build-image: ## Build the e2e-test image
	docker build -t $(REGISTRY)/e2e_test:$(TAG) .

.PHONY: publish-image
publish-image:
	docker push $(REGISTRY)/e2e_test:$(TAG)

.PHONY: e2e_test
e2e_test: check-go-version
	@CGO_ENABLED=0 go test -c -trimpath -ldflags="-buildid= -w" -o $@ .

.PHONY: clean
clean: ## Remove build artifacts
	$(RM_F) internal/pkg/assets/assets.go
	$(RM_F) $(PROGRAMS)

.PHONY: codegen
codegen: check-go-version ## Generate or update missing Go code defined in feature files
	@go run hack/codegen.go -update -dest-path=steps/conformance features/conformance

.PHONY: verify-codegen
verify-codegen: check-go-version ## Verify if generated Go code is in sync with feature files
	@go run hack/codegen.go -dest-path=steps/conformance features

.PHONY: verify-gherkin
verify-gherkin: check-go-version ## Verify format of gherkin feature files
	@hack/verify-gherkin.sh

.PHONY: help
help: ## Display this help
	@echo Targets:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9._-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: check-go-version
check-go-version:
	@hack/check-go-version.sh
