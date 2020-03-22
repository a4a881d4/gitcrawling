import shutil
import sys
import os

for i in os.listdir(sys.argv[1]):
    os.mkdir(os.path.join(sys.argv[1],"temp"))
    shutil.move(os.path.join(sys.argv[1],i),os.path.join(sys.argv[1],"temp",i))
    shutil.move(os.path.join(sys.argv[1],"temp"),os.path.join(sys.argv[1],i[:2]))
