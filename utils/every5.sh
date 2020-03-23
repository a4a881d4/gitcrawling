#! /bin/bash
PROCESS_NUM=`ps -ef | grep pclone | grep -v "grep" | grep -v "checkProcess.sh" | wc -l`
echo $PROCESS_NUM
if [ $PROCESS_NUM -eq 0 ];
then
    cd /home/pi/works/sda/gitdb
    wc -l ./task/$1/miss/miss
    python3 ./bin/clearup.py ./task/$1/miss/miss
    nohup ./bin/pclone -m ./task/$1/miss -t 16 > $1.log &
fi
