#! /usr/bin/python
import json
import sys
import shutil
import os

fo = open(sys.argv[1])
for line in fo.readlines():
    line = line.strip().split("/")
    if len(line[0])>2:
        dir = sys.argv[2]+"/packs/"+line[0][:2]
    else:
        dir = sys.argv[2]+"/packs/"+line[0]
    dir += "/"+"/".join(line)
    if os.path.exists(dir):
        if os.path.exists(dir+"/tmp-pack"):
            print(dir,"need remove")
            shutil.rmtree(dir)
        