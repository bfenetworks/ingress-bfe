#!/bin/sh

mkdir /bfe/log

STD_ERR_LOG="/bfe/log/stderr.log"
STD_OUT_LOG="/bfe/log/stdout.log"



cd /bfe/bin/ && ./bfe -c ../conf -l ../log 2>>$STD_ERR_LOG 1>>$STD_OUT_LOG &
#
cd /bfe/bin/ && ./bfe_ingress_controller -l ../log -u "http://localhost:8421/reload/" -c "/bfe/conf/" $@ 2>>$STD_ERR_LOG 1>>$STD_OUT_LOG &

fail_count=0
max_fail_count=6

while true;do
  if [ "${fail_count}"x = "$max_fail_count"x ];then
    exit 1
  fi

  bfe_proc=$(ps aux | grep bfe| grep -v ingress | grep -v grep | wc -l)
  bfe_ingress_proc=$(ps aux | grep bfe_ingress_controller | grep -v grep | wc -l)
  if [ "$bfe_proc"x != "1x" ]||[ "$bfe_ingress_proc"x != "1x" ];then
    fail_count=$(($fail_count+1))
  else
    fail_count=0
  fi

  sleep 10s
done