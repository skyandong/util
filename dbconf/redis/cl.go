package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

type clusterEntry struct {
	// Nodes address
	Nodes []string `mapstructure:"nodes"`
	// Auth password
	Auth string `mapstructure:"auth"`
	// PoolSize just as is
	PoolSize int `mapstructure:"poolSize"`
	// KeepAlive connection count
	KeepAlive int `mapstructure:"keepAlive"`
	// ConnTimeout just as is
	ConnTimeout time.Duration `mapstructure:"connTimeout"`
	// ReadTimeout just as is
	ReadTimeout time.Duration `mapstructure:"readTimeout"`
	// WriteTimeout just as is
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	// AliveTime for a connection
	AliveTime time.Duration `mapstructure:"aliveTime"`
	cluster   *redis.ClusterClient
	init      sync.Once
}

// ClusterConf for redis cluster map
type ClusterConf map[string]*clusterEntry

// Get cluster by name
func (c ClusterConf) Get(name string) *redis.ClusterClient {
	et, ok := c[name]
	if !ok {
		return nil
	}
	et.init.Do(func() {
		et.cluster = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        et.Nodes,
			DialTimeout:  et.ConnTimeout,
			ReadTimeout:  et.ReadTimeout,
			WriteTimeout: et.WriteTimeout,

			Password: et.Auth,
		})
	})
	return et.cluster
}

// Ensure exist and reachable
func (c ClusterConf) Ensure(names []string) error {
	for _, name := range names {
		cluster := c.Get(name)
		if cluster == nil {
			return fmt.Errorf("redis cluster %s not exist", name)
		}
		_, err := cluster.Ping(context.Background()).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

// Names returns keys in conf
func (c ClusterConf) Names() []string {
	names := make([]string, 0, len(c))
	for key := range c {
		names = append(names, key)
	}
	return names
}
