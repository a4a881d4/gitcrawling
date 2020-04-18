import sys
import shutil
import os

for root, dirs, files in os.walk(sys.argv[1], topdown=False):
    for name in files:
        if "pack-" in name and (".pack" in name or ".idx" in name):
            fn = os.path.join(root, name)
            shutil.copy(fn,sys.argv[2])
            print(fn)
            