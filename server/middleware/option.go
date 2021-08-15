package middleware

import "time"

// options for interceptor
type options struct {
	// for error log
	latencyLimit time.Duration
	// for full chain trace
	serviceName string
	// for trace log
	logFilter func(m []interface{})
	// for header log
	headerPicker func(k string) bool
	// for path filter
	pathFilter []string
}

// Option for interceptor
type Option func(o *options)

// LatencyLimit for error log
func LatencyLimit(limit time.Duration) Option {
	return func(o *options) {
		o.latencyLimit = limit
	}
}

// ServiceName for full chain trace
func ServiceName(sn string) Option {
	return func(o *options) {
		o.serviceName = sn
	}
}

// LogFilter for trace log
func LogFilter(fn func(m []interface{})) Option {
	return func(o *options) {
		o.logFilter = fn
	}
}

// HeaderPicker for log
func HeaderPicker(fn func(k string) bool) Option {
	return func(o *options) {
		o.headerPicker = fn
	}
}

// PathFilter for trace log
func PathFilter(f []string) Option {
	return func(o *options) {
		o.pathFilter = f
	}
}
