#!/bin/sh
# Copyright 2021 The BFE Authors
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
#
set -e
set -x

WORK_ROOT="$(cd "$(dirname "$0")/.." && pwd -P)"

# init version
VERSION=$(cat $WORK_ROOT/VERSION)
# init git commit id
GIT_COMMIT=$(git rev-parse HEAD) || true

go build -ldflags "-X main.version=${VERSION} -X main.commit=${GIT_COMMIT}" \
	-o $WORK_ROOT/output/bfe-ingress-controller $WORK_ROOT/cmd/ingress-controller

# set permission for docker
cp $WORK_ROOT/scripts/* $WORK_ROOT/output/
chmod a+x $WORK_ROOT/output/*
echo "${GIT_COMMIT}" > $WORK_ROOT/output/ingress.commit
