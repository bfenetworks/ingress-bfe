#!/bin/bash
set -x

# origin BFE root in docker
# see https://github.com/bfenetworks/bfe/blob/develop/Dockerfile
DOCKER_BFE_ROOT=/bfe
# new BFE root for ingress
WORK_BFE_ROOT=/home/work/go-bfe

mkdir -p ${WORK_BFE_ROOT}
cp -r ${DOCKER_BFE_ROOT}/bin ${WORK_BFE_ROOT}/
cp -r ${DOCKER_BFE_ROOT}/conf ${WORK_BFE_ROOT}/

# create directory for cert files
mkdir -p "${WORK_BFE_ROOT}/conf/tls_conf/certs"