#! /bin/bash
ps -ef|grep "uclone" | grep $1 |grep -v grep|cut -c 9-15|xargs kill -9
wc -l ./task/$1/miss/miss
python3 ./bin/clearpack.py ./task/$1/miss/miss $1
nohup ./bin/uclone -r $1 -m ./task/$1/miss -t 32 > $1.log &
