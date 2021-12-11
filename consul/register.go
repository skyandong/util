package consul

import (
	"strconv"

	"github.com/hashicorp/consul/api"
)

// Register http service
func (s *ServiceConf) Register() (err error) {
	fn := func(asr *api.AgentServiceRegistration) error {
		asr.Check.HTTP = "http://" + asr.Address + ":" + strconv.FormatInt(int64(s.Port), 10) + "/health"
		return nil
	}
	return s.register(fn)
}

// RegisterGRPC service
func (s *ServiceConf) RegisterGRPC() (err error) {
	fn := func(asr *api.AgentServiceRegistration) error {
		asr.Check.GRPC = asr.Address + ":" + strconv.FormatInt(int64(s.Port), 10) + "/" + asr.Name
		return nil
	}
	return s.register(fn)
}

// Deregister service
func (s *ServiceConf) Deregister() (err error) {
	agent, err := newAgent(s.Agent)
	if err != nil {
		return
	}
	return agent.ServiceDeregister(s.ID)
}

func (s *ServiceConf) register(fn func(*api.AgentServiceRegistration) error) (err error) {
	asr, err := s.prepare()
	if err != nil {
		return
	}
	err = fn(asr)
	if err != nil {
		return
	}
	agent, err := newAgent(s.Agent)
	if err != nil {
		return
	}
	return agent.ServiceRegister(asr)
}

func newAgent(c *AgentConf) (agent *api.Agent, err error) {
	config := api.DefaultConfig()
	if c != nil {
		config.Address = c.Address
	}
	client, err := api.NewClient(config)
	if err != nil {
		return
	}
	agent = client.Agent()
	return
}
