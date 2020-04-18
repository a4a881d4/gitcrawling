import leveldb as db
import os,sys,shutil

def parsekv(root,text):
    rs = root.split('/')
    ts = text.split(' ')
    tss = ts[0][:-1].split('/')
    hash = ts[1][:40]
    rs = rs[-2:]+tss[2:]
    return "/".join(rs),hash

def uclone2ref(dbf,rpath):
    rdb = db.LevelDB(dbf)
    batch = db.WriteBatch()
    def doFile(root, name):
        fn = os.path.join(root, name)
        with open(fn,'r') as file:
            for text in file.readlines():
                ref,h = parsekv(root,text)
                batch.Put(str.encode('refs/'+h),str.encode(ref))
                batch.Put(str.encode('proj/'+ref),str.encode(h))
                print(ref,h)
    # def doFile(root, name):
    #     fn = os.path.join(root, name)
    #     with open(fn,'r') as file:
    #         for text in file.readlines():
    #             ref,h = parsekv(root,text)
    #             batch.Put(str.encode('refs/'+h),str.encode(ref))
    #             batch.Put(str.encode('proj/'+ref),str.encode(h))
    #             print(ref,h)
    for root, dirs, files in os.walk(rpath, topdown=False):
        for name in files:
            if "refs" == name:
                doFile(root,name)
        # for di in dirs:
        #     if "refs" == di:
        #         doDir(root,di)
    rdb.Write(batch)

def dump(dbf):
    rdb = db.LevelDB(dbf)
    keys_values = list(rdb.RangeIter(key_from=str.encode('refs/')))
    for k,v in keys_values:
        p = rdb.Get(str.encode("proj/"+v.decode()))
        if k[5:]==p:
            print("hit:",p.decode(),v.decode())    

if __name__ == "__main__":
    uclone2ref(sys.argv[1],sys.argv[2])
    dump(sys.argv[1])