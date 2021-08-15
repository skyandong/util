package http

import (
	"github.com/skyandong/tool/service"
	"net/http"
	"sync"
	"time"
)

// ClientCache for HTTP
type ClientCache struct {
	cache       map[string]*http.Client
	lock        sync.RWMutex
	maxIdleConn int
	idleTimeout time.Duration
}

// DefaultClientCache ...
var DefaultClientCache *ClientCache

func init() {
	DefaultClientCache = NewClientCache(50, 5*time.Second)
}

// Get a HTTP client form global cache
func Get(service Service) *http.Client {
	return DefaultClientCache.Get(service)
}

// NewClientCache create a new cache
func NewClientCache(maxIdleConn int, idleTimeout time.Duration) *ClientCache {
	return &ClientCache{
		cache:       make(map[string]*http.Client),
		maxIdleConn: maxIdleConn,
		idleTimeout: idleTimeout,
	}
}

// Get a HTTP client
func (m *ClientCache) Get(s Service) *http.Client {
	// service key
	key := service.Service(s).String()

	// get from cache
	m.lock.RLock()
	c, ok := m.cache[key]
	m.lock.RUnlock()

	// not found
	if !ok {
		m.lock.Lock()

		// try get again
		c, ok = m.cache[key]
		if !ok {
			c = &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:    m.maxIdleConn,
					IdleConnTimeout: m.idleTimeout,
				},
			}
			m.cache[key] = c
		}
		m.lock.Unlock()
	}
	return c
}
