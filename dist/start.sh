#!/bin/sh
set -x

readonly BFE_BIN=bfe

cd /home/work/bfe/bin/ && nohup ./${BFE_BIN} -c ../conf -l ../log -d &

if [ -n "$INGRESS_LISTEN_NAMESPACE" ]; then
  cd /home/work/bfe/bin/ && ./bfe_ingress_controller -l ../log -c "/home/work/bfe/conf/" -n "$INGRESS_LISTEN_NAMESPACE" "$@"
else
  cd /home/work/bfe/bin/ && ./bfe_ingress_controller -l ../log -c "/home/work/bfe/conf/" "$@"
fi