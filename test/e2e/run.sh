#!/usr/bin/env bash

# Copyright 2020 The BFE Authors.
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

set -ex

trap TERM

cd "$(dirname "$0")"

RESULTS_DIR="${RESULTS_DIR:-$PWD/result}"
if [[ -d $RESULTS_DIR ]]; then
    rm -rf $RESULTS_DIR/*
else
    mkdir $RESULTS_DIR
fi

make build

CUCUMBER_OUTPUT_FORMAT="${CUCUMBER_OUTPUT_FORMAT:-pretty}"
WAIT_FOR_STATUS_TIMEOUT="${WAIT_FOR_STATUS_TIMEOUT:-5m}"
TEST_TIMEOUT="${TEST_TIMEOUT:-0}"
TEST_PARALLEL="${TEST_PARALLEL:-5}"

./e2e_test \
    --output-directory="${RESULTS_DIR}" \
    --feature="${CUCUMBER_FEATURE}" \
    --format="${CUCUMBER_OUTPUT_FORMAT}" \
    --wait-time-for-ingress-status="${WAIT_FOR_STATUS_TIMEOUT}" \
    --wait-time-for-ready="${WAIT_FOR_STATUS_TIMEOUT}" \
    --test.timeout="${TEST_TIMEOUT}" \
    --feature-parallel="${TEST_PARALLEL}"
ret=$?

exit 0
