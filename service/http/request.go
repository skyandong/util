package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/skyandong/util/service"
	"github.com/skyandong/util/trace"
)

func newJSONRequest(ctx context.Context, method, url string, param interface{}) (*http.Request, error) {
	var body io.Reader
	if param != nil {
		data, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if tid := service.GetTraceID(ctx); tid != "" {
		req.Header.Set(service.HeaderTraceID, tid)
	}
	for k, v := range service.GetExtraHeaders(ctx) {
		req.Header.Set(k, v)
	}
	span := trace.SpanFromContext(ctx)
	if span != nil {
		trace.InfoToHeader(req.Header, span.TraceInfo)
	}
	return req, nil
}
