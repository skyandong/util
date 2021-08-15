package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/skyandong/tool/native"
	"github.com/skyandong/tool/service"
	"github.com/skyandong/tool/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

//LocalIP 本机IP
var LocalIP string

func init() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Oops, InterfaceAddrs err: %v", err)
	}
	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				LocalIP = ipNet.IP.String()
				return
			}
		}
	}
}

/*
{
    trace: {
        from:  // trace来源，如某个业务的某个模块
        step:  // trace深度，请求到第几层了
        id:    // trace id，每次请求都有的一个唯一的id，比如可以采用每次生成不重复的UUID
    },
    addr: {
        rmt:  // 对端的ip和端口，比如10.2.3.4:8000
        loc:  // 本地的ip和端口，比如10.2.3.5:8020
    },
    time: // 毫秒 201809221830100 「global」
    params:{} // 请求参数
    elapse: // 耗时，毫秒
    result: { // 返回结果
        ret: 1,
        data: {}
    },
    level: //INFO、ERROR、WARNING、DEBUG 「global」
    path: “filename.go:300:/favor/add", // 打点位置  「global」
    ext: {} // 自定义字段。不超过5个
}
*/

type addr struct {
	RemoteIP string
	LocalIP  string
}

func (a *addr) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("rmt", a.RemoteIP)
	enc.AddString("loc", a.LocalIP)
	return nil
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Boss Web中间件
func Boss(logger *zap.SugaredLogger, limit time.Duration) gin.HandlerFunc {
	return BossWithOptions(logger, LatencyLimit(limit))
}

// BossWithOptions for gin
func BossWithOptions(logger *zap.SugaredLogger, ops ...Option) gin.HandlerFunc {
	options := &options{
		latencyLimit: 300 * time.Millisecond,
		serviceName:  LocalIP,
	}
	for _, opt := range ops {
		opt(options)
	}
	skips := map[string]struct{}{"/health": {}, "/metrics": {}}
	for _, path := range options.pathFilter {
		skips[path] = struct{}{}
	}
	return func(c *gin.Context) {
		traceID := c.Request.Header.Get(service.HeaderTraceID)
		if traceID == "" {
			traceID = uuid.NewV4().String()
		}
		c.Set(string(service.TraceID), traceID)

		start := time.Now()
		URLPath := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if _, skip := skips[URLPath]; skip {
			c.Next()
			return
		}
		if raw != "" {
			URLPath = URLPath + "?" + raw
		}
		clientIP := c.ClientIP()
		fullPath := c.FullPath()

		reqBuf, _ := ioutil.ReadAll(c.Request.Body)
		reqBuf1 := ioutil.NopCloser(bytes.NewBuffer(reqBuf))
		reqBuf2 := ioutil.NopCloser(bytes.NewBuffer(reqBuf))
		c.Request.Body = reqBuf2
		reqBody, _ := ioutil.ReadAll(reqBuf1)

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// full chain trace
		ap := &trace.AddressPair{
			RemoteIP: clientIP,
			LocalIP:  LocalIP,
		}
		ti := trace.InfoFromHeader(c.Request.Header, options.serviceName, start.UnixNano())
		span := trace.NewSpan(ti, ap, URLPath, logger)
		c.Set(string(trace.FullTraceSpanKey), span)

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		_addr := &addr{
			RemoteIP: clientIP,
			LocalIP:  LocalIP,
		}

		elapse := latency.Nanoseconds() / int64(time.Millisecond)
		status := c.Writer.Status()
		m := []interface{}{
			"addr", _addr,
			"elapse", elapse,
			"url", fullPath,
			"status", status,
			"method", c.Request.Method,
			"body", processBinary(reqBody),
			"response", processBinary(blw.body.Bytes()),
			"traceID", traceID,
			"traceInfo", span.TraceInfo,
		}
		if h := pickHeaders(c.Request.Header, options.headerPicker); len(h) > 0 {
			m = append(m, "headers", h)
		}
		if URLPath != fullPath {
			m = append(m, "path", URLPath)
		}
		if options.logFilter != nil {
			options.logFilter(m)
		}
		logger.Infow("jaeger-trace", m...)
		if status != http.StatusOK || elapse >= int64(options.latencyLimit/time.Millisecond) {
			logger.Errorw("error-trace", m...)
		}
	}
}

func pickHeaders(headers http.Header, picker func(string) bool) json.RawMessage {
	if len(headers) <= 0 || picker == nil {
		return nil
	}
	s := make([]string, 0, len(headers))
	for k, vs := range headers {
		if !picker(k) {
			continue
		}
		for _, v := range vs {
			s = append(s, k+": "+v)
		}
	}
	if len(s) <= 0 {
		return nil
	}
	d, _ := json.Marshal(s)
	return d
}

func processBinary(data []byte) interface{} {
	l := len(data)
	// be compatible with json.RawMessage
	if l <= 0 {
		return nil
	}
	// json object
	if l >= 2 && (data[0] == '{' || data[0] == '[') {
		return json.RawMessage(data)
	}
	return native.ByteSliceToString(data)
}
