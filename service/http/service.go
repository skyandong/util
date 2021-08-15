package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/skyandong/tool/service"
	"github.com/skyandong/tool/service/http/namecli"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// Consul 基于consul
	Consul = "consul"
	// NameSrv 基于namesrv
	NameSrv = "namesrv"
	// Origin 或 "" 为原始域名或IP
	Origin = "origin"
	// Secure https
	Secure = "secure"
)

// Service 定义一个服务
type Service service.Service

var localIP string

func init() {
	localIP = "127.0.0.1"
	k8sNodeIP := os.Getenv("K8S_NODE_IP")
	k8sNodeIP = strings.TrimSpace(k8sNodeIP)
	if k8sNodeIP != "" {
		localIP = k8sNodeIP
	}
	service.RegisterConverter(Consul, callHTTP)
	service.RegisterConverter(NameSrv, callHTTP)
	service.RegisterConverter(Origin, callHTTP)
	service.RegisterConverter(Secure, callHTTP)
	service.RegisterConverter("", callHTTP)
}

func callHTTP(ctx context.Context, svc service.Service, path string, req, reply interface{}) (err error) {
	data, err := Service(svc).PostJSON(ctx, path, req)
	if err != nil {
		return
	}
	return json.Unmarshal(data, reply)
}

// Call 调用服务并解析结果到指定类型
func (s Service) Call(ctx context.Context, path string, req, reply interface{}) (err error) {
	data, err := s.PostJSON(ctx, path, req)
	if err != nil {
		return
	}
	return json.Unmarshal(data, reply)
}

// URL 返回path对应的url
func (s Service) URL(ctx context.Context, path string) (string, error) {
	schema := "http://"
	if s.Type == Secure {
		schema = "https://"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	switch s.Type {
	case Consul:
		return strings.Join([]string{schema, localIP + ":9090/", s.Name, path}, ""), nil
	case NameSrv:
		addr, err := namecli.Name(ctx, s.Name)
		if err != nil {
			return "", err
		}
		return strings.Join([]string{schema, addr, path}, ""), nil
	case Secure, Origin, "":
		return strings.Join([]string{schema, s.Name, path}, ""), nil
	}
	return "", service.ErrServiceType
}

// GetJSON 请求对应的服务并返回原始结果数据
func (s Service) GetJSON(ctx context.Context, path string) (data []byte, err error) {
	ts := time.Now()
	data, err = s.requestJSON(ctx, http.MethodGet, path, nil)
	svc := service.Service(s)
	// be compatible with json.RawMessage
	if len(data) <= 0 {
		data = nil
	}
	svc.DoTrace(ctx, svc, path, nil, json.RawMessage(data), time.Now().Sub(ts), err)
	return
}

// PostJSON 请求对应的服务并返回原始结果数据
func (s Service) PostJSON(ctx context.Context, path string, param interface{}) (data []byte, err error) {
	ts := time.Now()
	data, err = s.requestJSON(ctx, http.MethodPost, path, param)
	svc := service.Service(s)
	// be compatible with json.RawMessage
	if len(data) <= 0 {
		data = nil
	}
	svc.DoTrace(ctx, svc, path, param, json.RawMessage(data), time.Now().Sub(ts), err)
	return
}

func (s Service) requestJSON(ctx context.Context, method, path string, param interface{}) (data []byte, err error) {
	url, err := s.URL(ctx, path)
	if err != nil {
		return nil, err
	}
	req, err := newJSONRequest(ctx, method, url, param)
	if err != nil {
		return nil, err
	}
	c := DefaultClientCache.Get(s)
	r, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := r.Body.Close(); err == nil {
			err = e
		}
	}()
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code: %d", r.StatusCode)
		return
	}
	return ioutil.ReadAll(r.Body)
}
