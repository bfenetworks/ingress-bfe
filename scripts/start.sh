#!/bin/sh
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
