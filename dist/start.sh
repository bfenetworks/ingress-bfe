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
set -x

readonly BFE_BIN=bfe

cd /home/work/bfe/bin/ && nohup ./${BFE_BIN} -c ../conf -l ../log -d &

if [ -n "$INGRESS_LISTEN_NAMESPACE" ]; then
  cd /home/work/bfe/bin/ && ./bfe_ingress_controller -l ../log -c "/home/work/bfe/conf/" -n "$INGRESS_LISTEN_NAMESPACE" "$@"
else
  cd /home/work/bfe/bin/ && ./bfe_ingress_controller -l ../log -c "/home/work/bfe/conf/" "$@"
fi