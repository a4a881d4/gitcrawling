import sys
import shutil
import os

def rename(fn):
    base = os.path.base(fn)
    