package namecli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	// Addr agent addr
	Addr string
	seq  uint32
)

func init() {
	ip := "127.0.0.1"
	k8sNodeIP := os.Getenv("K8S_NODE_IP")
	k8sNodeIP = strings.TrimSpace(k8sNodeIP)
	if k8sNodeIP != "" {
		ip = k8sNodeIP
	}
	Addr = ip + ":8328"
}

// Name resolve a service name to addr
func Name(ctx context.Context, name string) (addr string, err error) {
	if !strings.HasSuffix(name, ".ns") {
		return name, nil
	}

	timeout := time.Duration(0)
	retried := false
	deadline, ok := ctx.Deadline()
	if ok {
		timeout = deadline.Sub(time.Now())
	}
again:
	conn, err := net.DialTimeout("udp", Addr, timeout)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	err = conn.SetDeadline(deadline)
	if err != nil {
		return
	}

	seq := atomic.AddUint32(&seq, 1)
	req := fmt.Sprintf("%d,%s", seq, name)
	_, err = conn.Write([]byte(req))
	if err != nil {
		return
	}

	rsp := [64]byte{}
	n, err := conn.Read(rsp[:])
	if err != nil || n <= 0 {
		return
	}

	_rsp := rsp[:n]
	i := bytes.IndexByte(_rsp, ',')
	if i == -1 {
		// no expected
		return
	}
	_seq, _ := strconv.ParseUint(string(_rsp[:i]), 10, 32)
	if seq != uint32(_seq) {
		if !retried {
			retried = true
			goto again
		}
		err = errors.New("seq invalid from namesrv")
		return
	}

	addr = string(_rsp[i+1:])
	if addr == "" {
		if !retried {
			retried = true
			goto again
		}
		err = fmt.Errorf("no addr found from namesrv: %s", name)
		return
	}

	return
}
