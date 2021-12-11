package consul

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceConf for consul
type ServiceConf struct {
	// ID of service instance
	ID string
	// Name of service
	Name string
	// Address of service node
	Address string
	// Port of service instance
	Port int
	// Tags of service
	Tags []string
	// Meta for service
	Meta map[string]string
	// Check for service healthy
	Check *HealthCheckConf
	// Agent config for consul
	Agent *AgentConf
}

// HealthCheckConf for consul
type HealthCheckConf struct {
	// Interval for check
	Interval time.Duration
	// Timeout for check
	Timeout time.Duration
	// Critical deregister service
	Critical time.Duration
}

// AgentConf for consul
type AgentConf struct {
	// Address of agent
	Address string
}

const (
	defaultInterval = time.Second
	defaultTimeout  = 500 * time.Millisecond
)

var (
	idRegex   *regexp.Regexp
	nameRegex *regexp.Regexp
)

var (
	// ErrInvalidID in service config
	ErrInvalidID = errors.New("invalid service id")
	// ErrInvalidName in service config
	ErrInvalidName = errors.New("invalid service name")
	// ErrAddressRequired in service config
	ErrAddressRequired = errors.New("address is required")
	// ErrPortRequired in service config
	ErrPortRequired = errors.New("port is required")
)

func init() {
	idRegex = regexp.MustCompile(`^[A-Za-z][0-9A-Za-z]*(-[0-9A-Za-z]+)*(-[0-9]{1,3}(\.[0-9]{1,3}){3}(:[0-9]{2,5})?)$`)
	nameRegex = regexp.MustCompile("^[A-Za-z][0-9A-Za-z]*(-[0-9A-Za-z]+)*$")
}

func (s *ServiceConf) prepare() (asr *api.AgentServiceRegistration, err error) {
	err = s.validate()
	if err != nil {
		return
	}
	if s.ID == "" {
		s.ID = s.Name + "-" + s.Address + ":" + strconv.FormatInt(int64(s.Port), 10)
	}
	critical := time.Duration(0)
	interval := defaultInterval
	timeout := defaultTimeout
	if s.Check != nil {
		critical = s.Check.Critical
		interval = s.Check.Interval
		timeout = s.Check.Timeout
	}
	asr = &api.AgentServiceRegistration{
		ID:      s.ID,
		Name:    s.Name,
		Address: s.Address,
		Port:    s.Port,
		Tags:    s.Tags,
		Meta:    s.Meta,
		Check: &api.AgentServiceCheck{
			Interval: interval.String(),
			Timeout:  timeout.String(),
		},
	}
	if critical != 0 {
		asr.Check.DeregisterCriticalServiceAfter = critical.String()
	}
	return
}

func (s *ServiceConf) validate() error {
	if s.Name == "" || !nameRegex.MatchString(s.Name) {
		return ErrInvalidName
	}
	if s.ID != "" && !idRegex.MatchString(s.ID) {
		return ErrInvalidID
	}
	if s.Address == "" {
		return ErrAddressRequired
	}
	if s.Port == 0 {
		return ErrPortRequired
	}
	return nil
}
