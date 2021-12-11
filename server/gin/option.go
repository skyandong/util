package gin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skyandong/util/consul"
)

type options struct {
	serviceConfig []*consul.ServiceConf
	registerDelay time.Duration
	middleware    []gin.HandlerFunc
}

// Option for server
type Option func(*options)

// ServiceConf for consul
func ServiceConf(sc *consul.ServiceConf) Option {
	return func(o *options) {
		o.serviceConfig = []*consul.ServiceConf{sc}
	}
}

// ServiceConfigs for consul
func ServiceConfigs(cs []*consul.ServiceConf) Option {
	return func(o *options) {
		o.serviceConfig = cs
	}
}

// RegisterDelay duration
func RegisterDelay(rd time.Duration) Option {
	return func(o *options) {
		o.registerDelay = rd
	}
}

// Middleware for gin
func Middleware(ms ...gin.HandlerFunc) Option {
	return func(o *options) {
		o.middleware = ms
	}
}
