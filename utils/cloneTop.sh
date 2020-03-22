#! /bin/bash
wc -l ./task/$1/miss/miss
nohup ./bin/bigclone -m ./task/$1/miss -t 32 > $1.log &