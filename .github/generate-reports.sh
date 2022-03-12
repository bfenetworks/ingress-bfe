#!/bin/bash

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

set -o errexit
set -o nounset
set -o pipefail
set -x

(cd "$(dirname "$0")/../test/e2e/images/reports" && make build-image)

REPORT_BUILDER_IMAGE=local/reports-builder:0.0.1

REPORTS_DIR=${REPORTS_DIR:-/tmp/bfe-ingress-reports}

INGRESS_CONTROLLER="BFE-ingress-controller"
CONTROLLER_VERSION=${CONTROLLER_VERSION:-'N/A'}

TEMP_CONTENT=$(mktemp -d)

docker run \
  -e BUILD="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
  -e INPUT_DIRECTORY=/input \
  -e OUTPUT_DIRECTORY=/output \
  -e INGRESS_CONTROLLER="${INGRESS_CONTROLLER}" \
  -e CONTROLLER_VERSION="${CONTROLLER_VERSION}" \
  -v "${REPORTS_DIR}":/input:ro \
  -v "${TEMP_CONTENT}":/output \
  -u "$(id -u):$(id -g)" \
  "${REPORT_BUILDER_IMAGE}"

pushd "${TEMP_WORKTREE}" > /dev/null

if [[ -d ./e2e-test ]]; then
    git rm -r ./e2e-test
else
    mkdir -p "${TEMP_WORKTREE}/e2e-test"
fi

# copy new content
cp -r -a "${TEMP_CONTENT}/." "${TEMP_WORKTREE}/e2e-test/"

# cleanup HTML
sudo apt-get install tidy
for html_file in e2e-test/*.html;do
  tidy -q --break-before-br no --tidy-mark no --show-warnings no --wrap 0 -indent -m "$html_file" || true
done

# configure git
git config --global user.email "action@github.com"
git config --global user.name "GitHub Action"
# commit changes
git add e2e-test
git commit -m "e2e test report"
git push --force --quiet

popd > /dev/null
