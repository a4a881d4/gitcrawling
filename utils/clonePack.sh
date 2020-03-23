#! /bin/bash
ps -ef|grep "uclone" |grep -v grep|cut -c 9-15|xargs kill -9
wc -l ./task/$1/miss/miss
python3 ./bin/clearpack.py ./task/$1/miss/miss
nohup ./bin/uclone -m ./task/$1/miss -t 32 > $1.log &
