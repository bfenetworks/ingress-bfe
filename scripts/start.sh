#!/bin/sh
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
#
set -e 
if [ $# -lt 2 ]; then
    echo "error: number of argument should >= 2"
    exit 1
fi

trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

CTL_BIN=$1
BFE_BIN=$2

BFE_NAME="$(basename $BFE_BIN)"
BFE_ROOT_DIR="$(cd "$(dirname "$BFE_BIN")/.." && pwd -P)"
CONF_DIR="$BFE_ROOT_DIR/conf"
BFE_BIN_DIR="$BFE_ROOT_DIR/bin"

shift 2
ARGS="$@ -c $CONF_DIR "
if [ -n "$INGRESS_LISTEN_NAMESPACE" ]; then
    ARGS=$ARGS"-n $INGRESS_LISTEN_NAMESPACE"
fi

cd ${BFE_BIN_DIR} && ./${BFE_NAME} -c ../conf -l ../log &
sleep 1
pgrep ${BFE_NAME}
if [ $? -ne 0 ]; then
     exit 1
fi

${CTL_BIN} $ARGS
