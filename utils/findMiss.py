import shutil
import sys
import json
import os

s = []

start = int(sys.argv[2])
end = int(sys.argv[3])

for i in range(start,end):
    fn = sys.argv[1]+"{:0>6d}".format(i)+".star"
    f = open(fn)
    star = json.load(f)
    f.close()

    for k in star:
        s += star[k]

for n in s:
    print n