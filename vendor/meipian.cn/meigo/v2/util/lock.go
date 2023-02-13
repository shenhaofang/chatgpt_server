package util

import (
	"sync"
	"time"

	"meipian.cn/meigo/v2/config"
	"meipian.cn/meigo/v2/log"

	"github.com/gomodule/redigo/redis"
	"gopkg.in/redsync.v1"
)

var lockPool []redsync.Pool
var lockMu sync.Mutex

// Lock 分布锁
func Lock(name string, options ...LockOption) (mu *redsync.Mutex, code int) {

	initLock()

	rd := redsync.New(lockPool)
	mu = rd.NewMutex(name, options...)
	err := mu.Lock()
	if err != nil {
		log.Err(err.Error())
		code = ErrLock
		return
	}

	return
}

func initLock() {

	lockMu.Lock()
	defer lockMu.Unlock()

	if len(lockPool) == 0 {
		cfg := config.RedisConfig("redis")
		item := &redis.Pool{
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", cfg.Host, redis.DialPassword(cfg.Auth))
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
		}

		lockPool = append(lockPool, item)
	}
}
