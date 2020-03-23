#! /bin/bash
ps -ef|grep "pclone" |grep -v grep|cut -c 9-15|xargs kill -9
wc -l ./task/$1/miss/miss
python3 ./bin/clearup.py ./task/$1/miss/miss
nohup ./bin/pclone -m ./task/$1/miss -t 16 > $1.log &
