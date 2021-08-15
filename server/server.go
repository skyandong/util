package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/skyandong/tool/consul"
	"github.com/skyandong/tool/program_controller"
	"log"
	"net"
	"net/http"
	"time"
)

// Server for gin service
type Server struct {
	rd time.Duration
	sc []*consul.ServiceConf
	hs *http.Server
}

var _ program_controller.Server = (*Server)(nil)

// NewServer creates a gin server
func NewServer(ops ...Option) *Server {
	o := options{
		registerDelay: 2 * time.Second,
	}
	for _, op := range ops {
		op(&o)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	for _, m := range o.middleware {
		engine.Use(m)
	}
	engine.Any("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "SUCCESS")
	})
	s := &Server{
		rd: o.registerDelay,
		sc: o.serviceConfig,
		hs: &http.Server{
			Handler: engine,
		},
	}
	return s
}

// Serve on the listener
func (s *Server) Serve(l net.Listener) (err error) {
	var over bool
	defer func() {
		over = true
		// normal close, already deregister
		if err == nil {
			return
		}
		// deregister after serve error
		if e := s.deregister(); e != nil {
			log.Printf("deregister service error: %v", e)
		}
	}()
	go func() {
		time.Sleep(s.rd)
		if over {
			return
		}
		// delay register, in case serve fails
		if e := s.register(); e != nil {
			log.Printf("register service error: %v", e)
		}
	}()
	err = s.hs.Serve(l)
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	if e := s.deregister(); e != nil {
		log.Printf("deregister service error: %v", e)
	}
	return s.hs.Shutdown(ctx)
}

// Origin gin engine
func (s *Server) Origin() *gin.Engine {
	return s.hs.Handler.(*gin.Engine)
}

func (s *Server) register() error {
	for _, c := range s.sc {
		if err := c.Register(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) deregister() error {
	for _, c := range s.sc {
		if err := c.Deregister(); err != nil {
			return err
		}
	}
	return nil
}
