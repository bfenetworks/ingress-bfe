#!/bin/sh
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
