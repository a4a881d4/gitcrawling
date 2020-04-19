#! /usr/bin/python
import json
import sys

f = open(sys.argv[1])
star = json.load(f)
f.close()
s = 0
for k in star:
    s += len(star[k])
print(s)
