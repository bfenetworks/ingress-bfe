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
set -e

download_kubectl(){
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # linux
        echo "linux"
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac
        if [[ $(arch) == 'arm64' ]]; then
            curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/arm64/kubectl"

        else
            curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
        fi
    else
        echo "unsupported os type: " "$OSTYPE"
        exit 1
    fi

    chmod +x ./kubectl

}

cd "$(dirname "$0")"
VERSION=$1

IMAGE="bfenetworks/bfe-ingress-controller:"$VERSION

if [[ "$(docker images -q $IMAGE 2> /dev/null)" == "" ]]; then
    echo "image does not exist:" "$IMAGE"
    exit 1
fi

if [[ ! -f kubectl ]]; then
    download_kubectl
fi

# update yaml to version
sed "s#image: .*\$#image: $IMAGE#g" ../../examples/controller-all.yaml > controller-all.yaml

./kubectl apply -f controller-all.yaml
./kubectl apply -f controller-svc.yaml -f ingressclass.yaml

