import os
import sys
import shutil

def listdir(path):
    base = os.path.basename(path)
    for o in os.listdir(path):
        op = os.path.join(path,o)
        if o[:2]==base:
            listdir(op)
        else:
            if os.path.isdir(op):
                print(base+"/"+o) 

for o in os.listdir(sys.argv[1]):
    op = os.path.join(sys.argv[1],o)    
    listdir(op)