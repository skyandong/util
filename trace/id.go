package trace

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var pool *sync.Pool

func init() {
	rand.Seed(time.Now().UnixNano())
	pool = &sync.Pool{
		New: func() interface{} {
			return rand.NewSource(rand.Int63())
		},
	}
}

// GetJaegerTraceID for full chain trace
func GetJaegerTraceID() string {
	generator := pool.Get().(rand.Source)
	number := uint64(generator.Int63())
	pool.Put(generator)

	id := strconv.FormatUint(number, 16)
	return id
}
