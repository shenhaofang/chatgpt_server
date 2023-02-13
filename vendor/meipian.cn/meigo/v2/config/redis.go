package config

import (
	"fmt"
	"sync"
)

var redisM sync.Map

type Redis struct {
	Host     string
	Auth     string
	PoolSize int
	Prefix   string // key前缀
}

const redisDftPoolSize = 10

// RedisConfig 获取redis配置
func RedisConfig(conn string) (redis Redis) {
	if ri, ok := redisM.Load(conn); ok {
		return ri.(Redis)
	}

	redis.Host = GetDft(fmt.Sprintf("%s.host", conn), "127.0.0.1:6379")
	redis.Auth = GetStr(fmt.Sprintf("%s.auth", conn))
	redis.PoolSize = GetIntDft(fmt.Sprintf("%s.pool_size", conn), redisDftPoolSize)
	redis.Prefix = GetStr(fmt.Sprintf("%s.prefix", conn))

	redisM.Store(conn, redis)
	return
}
