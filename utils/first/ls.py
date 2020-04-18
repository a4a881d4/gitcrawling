import leveldb as db
import sys,json

def parsekv(k,ref):
    rs = k.split('/')
    tss = ref['Name'].split('/')
    rs = rs[-2:]+tss[2:]
    return "/".join(rs),ref['Hash']

odb = db.LevelDB(sys.argv[1])

for k,v in odb.RangeIter(key_from=str.encode('r/')):
    vo = json.loads(v)
    for ref in vo['Refs']:
        ref,h = parsekv(k.decode(),ref)
        
        print(ref,h)