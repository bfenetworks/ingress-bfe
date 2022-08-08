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
FROM golang:1.16-alpine3.14 AS build

RUN apk add build-base

WORKDIR /bfe-ingress-controller
COPY . .
RUN build/build.sh

FROM bfenetworks/bfe:v-1.3.0
WORKDIR /
COPY --from=build /bfe-ingress-controller/output/* /

EXPOSE 8080 8443 8421

ENTRYPOINT ["/bfe-ingress-controller"]
