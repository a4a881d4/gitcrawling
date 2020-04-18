#! /usr/bin/python
import shutil
import sys
import json
import os

file = open(sys.argv[1])
projects = {}
for f in file.readlines():
    f = f.strip()
    # print f
    projects[f] = 1

file.close()
total = len(projects)
print "Total",total
repos = []
packs = []
bads = []
def doRepos(path):
    def listrepos(path):
        base = os.path.basename(path)
        for o in os.listdir(path):
            op = os.path.join(path,o)
            # if o[:2].lower()==base.lower():
            #     listrepos(op)
            # else:
            find = False
            if os.path.isdir(op) and os.path.isdir(os.path.join(op,"objects","pack")):
                packfiles = os.listdir(os.path.join(op,"objects","pack"))
                for pf in packfiles:
                    if "pack-" in pf:
                        if base+"/"+o in projects:
                            del projects[base+"/"+o]
                            find = True
            if not find:
                bads.append(base+"/"+o)
                print "bad",base+"/"+o
    
    for o in os.listdir(path):
        op = os.path.join(path,o)    
        listrepos(op)


def listdir(path):
    dirs = os.listdir(path)
    for f in dirs:
        if f=="repos":
            repos.append(path)
            return
        if f=="packs":
            packs.append(path)
            return
        chird = os.path.join(path,f)
        if os.path.isdir(chird):
            listdir(chird)

listdir(sys.argv[2])
print "Repos"
print repos
for r in repos:
    rpath = os.path.join(r,"repos")
    for fs in os.listdir(rpath):
        doRepos(os.path.join(rpath,fs))
print "Packs"
print packs
        
        
"""

for root,dirs,files in os.walk(sys.argv[2], topdown=True):
    for fn in files:
        if ".idx" in fn and "pack" in fn:
            root = root.replace("\\","/")
            ds = root.split("/")
            p = ds[-4]+"/"+ds[-3]
            projects.discard(p)
            total -= 1
            if total%100 == 0:
                print "*" 


"""
f = open(sys.argv[1],'w')
for k in projects:
    # print k
    print >> f,k
f.close()
