package trace

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

// Span for trace
type Span struct {
	Name      string
	Address   *AddressPair
	TraceInfo *SpanInfo
	Logger    *zap.SugaredLogger
}

// SpanInfo as trace info
type SpanInfo struct {
	TraceID      string  `json:"traceID"`
	StartTime    int64   `json:"startTime"`
	ServiceName  string  `json:"serviceName"`
	SpanID       string  `json:"spanID"`
	ParentSpanID string  `json:"pSpanId"`
	Sampler      string  `json:"sampler"`
	RefType      RefType `json:"refType"`
	Type         Type    `json:"type"`
}

// Type for span
type Type int

// RefType of parent span
type RefType int

// AddressPair for span
type AddressPair struct {
	RemoteIP string `json:"rmt"`
	LocalIP  string `json:"loc"`
}

const (
	// XB3TraceID for the chain
	XB3TraceID = "x-b3-traceid"
	// XB3ParentSpanID for parent span
	XB3ParentSpanID = "x-b3-parentspanid"
	// XB3Sampled for sample rate
	XB3Sampled = "x-b3-sampled"
)

const (
	// TypeBossSpan for the boss span
	TypeBossSpan Type = iota + 1
	// TypeFinishSpan for user created span
	TypeFinishSpan
)

const (
	// RefTypeChildOf the parent span
	RefTypeChildOf RefType = iota
	// RefTypeFollowsFrom the parent span
	RefTypeFollowsFrom
)

const (
	// SampleBaseRate for sample
	SampleBaseRate = 100
)

// Finish the span, print log
func (s *Span) Finish() int64 {
	t := s.TraceInfo
	if t.TraceID == "" || t.Type == TypeBossSpan {
		return -1
	}
	timestamp := getTimestamp(0)
	elapse := (timestamp - t.StartTime) / int64(time.Millisecond/time.Microsecond)

	s.Logger.Infow("jaeger-trace",
		"addr", s.Address,
		"elapse", elapse,
		"url", s.Name,
		"traceInfo", t,
	)
	t.TraceID = ""

	return timestamp
}

// NewChild base on current span
func (s *Span) NewChild(name string) *Span {
	t := s.TraceInfo
	if t.TraceID == "" {
		return nil
	}
	i := &SpanInfo{
		TraceID:      t.TraceID,
		StartTime:    getTimestamp(0),
		ServiceName:  t.ServiceName,
		SpanID:       GetJaegerTraceID(),
		ParentSpanID: t.SpanID,
		Sampler:      t.Sampler,
		RefType:      RefTypeChildOf,
		Type:         TypeFinishSpan,
	}
	n := &Span{
		Name:      name,
		Address:   s.Address,
		TraceInfo: i,
		Logger:    s.Logger,
	}
	return n
}

// NewSubsequent create a subsequent span, and finish current span
func (s *Span) NewSubsequent(name string) *Span {
	t := s.TraceInfo
	if t.TraceID == "" {
		return nil
	}
	i := &SpanInfo{
		TraceID:      t.TraceID,
		ServiceName:  t.ServiceName,
		SpanID:       GetJaegerTraceID(),
		ParentSpanID: t.ParentSpanID,
		Sampler:      t.Sampler,
		RefType:      RefTypeChildOf,
		Type:         TypeFinishSpan,
	}
	a := s.Finish()
	if a <= 0 {
		a = getTimestamp(0)
	}
	i.StartTime = a
	n := &Span{
		Name:      name,
		Address:   s.Address,
		TraceInfo: i,
		Logger:    s.Logger,
	}
	return n
}

// NewFollow create a span follows from current, and finish current span
func (s *Span) NewFollow(name string) *Span {
	t := s.TraceInfo
	if t.TraceID == "" {
		return nil
	}
	i := &SpanInfo{
		TraceID:      t.TraceID,
		ServiceName:  t.ServiceName,
		SpanID:       GetJaegerTraceID(),
		ParentSpanID: t.SpanID,
		Sampler:      t.Sampler,
		RefType:      RefTypeFollowsFrom,
		Type:         TypeFinishSpan,
	}
	a := s.Finish()
	if a <= 0 {
		a = getTimestamp(0)
	}
	i.StartTime = a
	n := &Span{
		Name:      name,
		Address:   s.Address,
		TraceInfo: i,
		Logger:    s.Logger,
	}
	return n
}

// NewSpan create a span
func NewSpan(info *SpanInfo, addr *AddressPair, name string, logger *zap.SugaredLogger) *Span {
	return &Span{
		Name:      name,
		Address:   addr,
		TraceInfo: info,
		Logger:    logger,
	}
}

// InfoToContext append span info to context for grpc
func InfoToContext(ctx context.Context, info *SpanInfo) context.Context {
	kvs := []string{XB3TraceID, info.TraceID, XB3ParentSpanID, info.SpanID, XB3Sampled, info.Sampler}
	return metadata.AppendToOutgoingContext(ctx, kvs...)
}

// InfoToHeader set span info to http header
func InfoToHeader(hdr http.Header, info *SpanInfo) {
	hdr.Set(XB3TraceID, info.TraceID)
	hdr.Set(XB3ParentSpanID, info.SpanID)
	hdr.Set(XB3Sampled, info.Sampler)
}

// InfoFromMD extract span info from grpc metadata
func InfoFromMD(md metadata.MD, sn string, ts int64) *SpanInfo {
	info := &SpanInfo{
		TraceID:      getStringFromMD(md, XB3TraceID),
		StartTime:    getTimestamp(ts),
		ServiceName:  sn,
		ParentSpanID: getStringFromMD(md, XB3ParentSpanID),
		Sampler:      getStringFromMD(md, XB3Sampled),
	}
	initTraceSpanInfo(info)
	return info
}

// InfoFromHeader extract span info from http header
func InfoFromHeader(hdr http.Header, sn string, ts int64) *SpanInfo {
	info := &SpanInfo{
		TraceID:      getStringFromHeader(hdr, XB3TraceID),
		StartTime:    getTimestamp(ts),
		ServiceName:  sn,
		ParentSpanID: getStringFromHeader(hdr, XB3ParentSpanID),
		Sampler:      getStringFromHeader(hdr, XB3Sampled),
	}
	initTraceSpanInfo(info)
	return info
}

func initTraceSpanInfo(info *SpanInfo) {
	if info.TraceID == "" {
		info.TraceID = GetJaegerTraceID()
	}
	if info.Sampler == "" {
		info.Sampler = strconv.Itoa(rand.Intn(SampleBaseRate))
	}
	if info.ParentSpanID == "" {
		info.SpanID = info.TraceID
	} else {
		info.SpanID = GetJaegerTraceID()
		info.RefType = RefTypeChildOf
	}
	info.Type = TypeBossSpan
}

func getStringFromMD(md metadata.MD, key string) string {
	ts := md.Get(key)
	if len(ts) > 0 {
		return ts[0]
	}
	return ""
}

func getStringFromHeader(hdr http.Header, key string) string {
	return hdr.Get(key)
}

func getTimestamp(ts int64) int64 {
	if ts <= 0 {
		ts = time.Now().UnixNano()
	}
	return ts / int64(time.Microsecond)
}
