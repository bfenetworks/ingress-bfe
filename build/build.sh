#!/bin/bash



# init version
VERSION=$(cat VERSION)
# init git commit id
GIT_COMMIT=$(git rev-parse HEAD)

cd cmd/bfe_ingress_controller && GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION} -X main.commit=${GIT_COMMIT}" -o bfe_ingress_controller

cp bfe_ingress_controller ../../dist/ && rm bfe_ingress_controller
