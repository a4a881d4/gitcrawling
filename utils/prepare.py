import shutil
import sys
import json
import os

f = open(sys.argv[1])
star = json.load(f)
f.close()
s = []
for k in star:
    s += star[k]

c = 0
t = 0

os.mkdir("temp")
os.mkdir("temp/miss")
f = open("temp/miss/miss","w")

for n in s:
    print(n,file=f)
    c+=1
    if c>=40000:
        f.close()
        shutil.copy("temp/miss/miss","temp/origin")
        dir = sys.argv[2]+'.'+str(t)
        print("write to ",dir)
        shutil.move("temp",dir)
        t+=1
        os.mkdir("temp")
        os.mkdir("temp/miss")
        f = open("temp/miss/miss","w")
        c=0
f.close()
shutil.copy("temp/miss/miss","temp/origin")
dir = sys.argv[2]+'.'+str(t)
shutil.move("temp",dir)


