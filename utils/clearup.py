#! /usr/bin/python
import json
import sys
import shutil
import os

fo = open(sys.argv[1])
for line in fo.readlines():
    line = line.strip().split("/")
    if len(line[0])>2:
        dir = "repos/"+line[0][:2]
    else:
        dir = "repos/"+line[0]
    dir += "/"+"/".join(line)
    if os.path.exists(dir):
        pack = os.listdir(dir+"/objects/pack")
        if len(pack)==2:
            if ".idx" in pack[0] or ".pack" in pack[0]:
                continue
        print(dir,"need remove")
        shutil.rmtree(dir)
        