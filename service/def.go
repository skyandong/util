package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// HeaderKeyType for context.WithValue
type HeaderKeyType string

// TraceKeyType for context.WithValue
type TraceKeyType string

// CallerTraceFunc trace at caller side
type CallerTraceFunc func(ctx context.Context, svc Service, path string, req, reply interface{}, elapse time.Duration, err error)

// CallFunc for subclass
type CallFunc func(ctx context.Context, svc Service, path string, req, reply interface{}) (err error)

const (
	// ExtraHeaders for request
	ExtraHeaders HeaderKeyType = "extra-headers"
	// TraceID for key name
	TraceID TraceKeyType = "traceID"
	// HeaderTraceID for HTTP & GRPC
	HeaderTraceID = "trace-id"
)

var (
	// ErrServiceType error
	ErrServiceType = errors.New("service type mismatch")
	// converters to subclass
	converters = map[string]CallFunc{}
)

// Service 定义一个服务
type Service struct {
	// Type 服务发现类型 或 grpc
	Type string `json:"type" yaml:"type"`
	// Name 服务名
	Name string `json:"name" yaml:"name"`
	// Trace 调用者跟踪
	Trace CallerTraceFunc `json:"-" yaml:"-"`
}

// Call service
func (s Service) Call(ctx context.Context, path string, req, reply interface{}) (err error) {
	fn, ok := converters[s.Type]
	if !ok {
		return fmt.Errorf("unknown service type: %s", s.Type)
	}
	return fn(ctx, s, path, req, reply)
}

// DoTrace used by the middleware
func (s Service) DoTrace(ctx context.Context, svc Service, path string, req, reply interface{}, elapse time.Duration, err error) {
	if s.Trace != nil {
		if len(path) > 0 && path[0] == '/' && strings.HasPrefix(path[1:], s.Name) {
			path = strings.Join([]string{s.Type, path[1:]}, "://")
		} else {
			path = s.String() + path
		}
		s.Trace(ctx, svc, path, req, reply, elapse, err)
	}
}

// String 实现Stringer接口
func (s Service) String() string {
	return strings.Join([]string{s.Type, s.Name}, "://")
}

// RegisterConverter for subclass
func RegisterConverter(typ string, fn CallFunc) {
	converters[typ] = fn
}

// GetTraceID from context
func GetTraceID(c context.Context) string {
	if tid, ok := c.Value(TraceID).(string); ok {
		return tid
	}
	// be compatible with gin.Context
	if tid, ok := c.Value(string(TraceID)).(string); ok {
		return tid
	}
	return ""
}

// GetExtraHeaders from context
func GetExtraHeaders(c context.Context) map[string]string {
	if m, ok := c.Value(ExtraHeaders).(map[string]string); ok {
		return m
	}
	// be compatible with gin.Context
	if m, ok := c.Value(string(ExtraHeaders)).(map[string]string); ok {
		return m
	}
	return nil
}
