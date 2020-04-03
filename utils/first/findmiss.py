import os,sys,time
import leveldb as db
from github import Github

token = os.environ["GITHUB_TOKEN"]

print(token)

g = Github(login_or_token=token,per_page=200)
rdb = db.LevelDB(sys.argv[1])
fmiss = open(sys.argv[2],'w')
def test(fullname):
    kk = 'proj/'+fullname
    kp = str.encode(kk)
    for k,v in rdb.RangeIter(key_from=kp):
        if k[:len(kp)] == kp:
            return True
        else:
            return False
    return False

start = g.get_repo(sys.argv[3]) #"cartographer-project/cartographer"
stars = start.stargazers_count
qu = 'stars:>%d' % stars
print("begin",stars,qu)

for i in range(5):

    repositories = g.search_repositories(query=qu,sort='stars',order='asc').get_page(i)
    if len(repositories)==0:
        break
    for repo in repositories:
        if not test(repo.full_name):
            print(repo.full_name,file=fmiss)
            print("miss:",repo.full_name)
    # w,r = g.rate_limit 
    # while w<1:
    #     for i in range(r):
    #         time.sleep(1)
    #         print('*',end='')
fmiss.close()
