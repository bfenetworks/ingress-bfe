#!/usr/bin/env bash
# Copyright 2022 The BFE Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -ex


download_kind(){
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # linux
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64

    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac
        if [[ $(arch) == 'arm64' ]]; then
            curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-arm64
        else
            curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64
        fi
    else
        echo "unsupported os type: " "$OSTYPE"
        exit 1
    fi
    chmod +x ./kind
}

cd "$(dirname "$0")"

if [[ ! -f kind ]]; then
    download_kind
fi

# check if cluster exist
if ./kind get clusters | grep -Fxq "kind"; then
    exit 0
fi

./kind create cluster --config=./kind-config.yaml


