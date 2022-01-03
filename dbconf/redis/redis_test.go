package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var conf Conf

const (
	testRedis1 = "test_redis1"
	testRedis2 = "test_redis2"
)

func TestMain(m *testing.M) {
	conf = make(map[string]*entry)
	conf[testRedis1] = &entry{
		Address:        "localhost:6379",
		Password:       "testredisxixihaha123",
		Db:             0,
		MaxIdle:        32,
		PoolSize:       128,
		ConnectTimeout: time.Second,
		IdleTimeout:    time.Minute,
		ReadTimeout:    100 * time.Millisecond,
		WriteTimeout:   100 * time.Millisecond,
	}

	conf[testRedis2] = &entry{
		Address:        "localhost:6380",
		Password:       "testredisxixihaha123",
		Db:             0,
		MaxIdle:        32,
		PoolSize:       128,
		ConnectTimeout: time.Second,
		IdleTimeout:    time.Minute,
		ReadTimeout:    100 * time.Millisecond,
		WriteTimeout:   100 * time.Millisecond,
	}
	m.Run()
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	redis := conf.Get(testRedis1)
	assert.NotNil(redis)
}

func TestEnsure(t *testing.T) {
	require := require.New(t)

	err := conf.Ensure([]string{testRedis1, testRedis2})
	require.NoError(err)
}
