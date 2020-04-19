import os,sys,shutil

def parsekv(root,text):
    rs = root.split('/')
    ts = text.split(' ')
    tss = ts[0][:-1].split('/')
    hash = ts[1][:40]
    rs = rs[-2:]+tss[2:]
    return "/".join(rs),hash

def uclone2ref(rpath):
    def doFile(root, name):
        fn = os.path.join(root, name)
        with open(fn,'r') as file:
            for text in file.readlines():
                ref,h = parsekv(root,text)
                print(ref,h)
    for root, dirs, files in os.walk(rpath, topdown=False):
        for name in files:
            if "refs" == name:
                doFile(root,name)


if __name__ == "__main__":
    uclone2ref(sys.argv[1])
