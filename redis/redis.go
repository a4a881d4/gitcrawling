package redis

import (
	"github.com/gomodule/redigo/redis"
)

type DBClient struct {
	client redis.Conn
}

func Dial(url string) (*DBClient, error) {
	conn, err := redis.Dial("tcp", ":6379")
	return &DBClient{conn}, err
}

func(self *DBClient) Put(k,v []byte) {
	
}
