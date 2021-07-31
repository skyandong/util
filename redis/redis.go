package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"sync"
	"time"
)

type entry struct {
	// Address of redis
	Address string `mapstructure:"address"`
	// Password to auth
	Password string `mapstructure:"password"`
	// Db to select
	Db int `mapstructure:"db"`
	// MaxIdle connection count
	MaxIdle int `mapstructure:"maxIdle"`
	// PoolSize just as is
	PoolSize int `mapstructure:"poolSize"`
	// ConnectTimeout just as is
	ConnectTimeout time.Duration `mapstructure:"connectTimeout"`
	// IdleTimeout for an idle connection
	IdleTimeout time.Duration `mapstructure:"idleTimeout"`
	// ReadTimeout just as is
	ReadTimeout time.Duration `mapstructure:"readTimeout"`
	// WriteTimeout just as is
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	pool         *redis.Pool
	init         sync.Once
}

// Conf for redis map
type Conf map[string]*entry

// Get pool by name
func (c Conf) Get(name string) *redis.Pool {
	et, ok := c[name]
	if !ok {
		return nil
	}
	et.init.Do(func() {
		et.pool = &redis.Pool{
			MaxIdle:     et.MaxIdle,
			IdleTimeout: et.IdleTimeout,
			Dial: func() (conn redis.Conn, err error) {
				conn, err = redis.Dial("tcp", et.Address,
					redis.DialPassword(et.Password),
					redis.DialDatabase(et.Db),
					redis.DialConnectTimeout(et.ConnectTimeout),
					redis.DialReadTimeout(et.ReadTimeout),
					redis.DialWriteTimeout(et.WriteTimeout),
				)
				return
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < 5*time.Second {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
		}
	})
	return et.pool
}

// Ensure exist and reachable
func (c Conf) Ensure(names []string) error {
	for _, name := range names {
		pool := c.Get(name)
		if pool == nil {
			return fmt.Errorf("redis %s not exist", name)
		}
		conn := pool.Get()
		_, err := conn.Do("PING")
		_ = conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Names returns keys in conf
func (c Conf) Names() []string {
	names := make([]string, 0, len(c))
	for key := range c {
		names = append(names, key)
	}
	return names
}
