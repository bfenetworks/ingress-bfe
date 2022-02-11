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

cd "$(dirname "$0")"

# load bfe-ingress-controller image
VERSION=$1
IMAGE="bfenetworks/bfe-ingress-controller:"$VERSION

if [[ "$(docker images -q $IMAGE 2> /dev/null)" == "" ]]; then
    echo "image does not exist:" "$IMAGE"
    exit 1
fi

./kind load docker-image $IMAGE

# build and load backend image (echoserver)
IMAGE="local/echoserver:0.0.1"

if [[ "$(docker images -q $IMAGE 2> /dev/null)" == "" ]]; then
	(cd ../e2e/images/echoserver; make build-image)
fi

./kind load docker-image $IMAGE
