package cluster

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Zwlin98/moon/gate"
	"github.com/Zwlin98/moon/lua"
	"github.com/Zwlin98/moon/service"
)

type Clusterd interface {
	Reload(ClusterConfig)

	Register(any, service.Service) error
	Query(any) service.Service

	Open(string) error
	OnConnect(gate gate.Gate, conn net.Conn)

	fetchSender(string) Sender
	OnSenderExit(string)
}

type skynetClusterd struct {
	sync.Mutex

	config ClusterConfig

	namedServices map[string]service.Service

	nodeSender sync.Map

	gate map[string]gate.Gate
}

var globalClusterd Clusterd
var once sync.Once

func GetClusterd() Clusterd {
	once.Do(func() {
		globalClusterd = newClusterd()
	})
	return globalClusterd
}

func newClusterd() *skynetClusterd {
	return &skynetClusterd{
		namedServices: make(map[string]service.Service),
		gate:          make(map[string]gate.Gate),
		config:        make(DefaultConfig),
	}
}

func Call(node string, service string, method string, args []lua.Value) ([]lua.Value, error) {
	c := GetClusterd()
	client := c.fetchSender(node)
	if client == nil {
		return nil, fmt.Errorf("no client for node: %s", node)
	}
	return client.Call(service, method, args)
}

func Send(node string, service string, method string, args []lua.Value) error {
	c := GetClusterd()
	client := c.fetchSender(node)
	if client == nil {
		return fmt.Errorf("no client for node: %s", node)
	}
	return client.Send(service, method, args)
}

func (c *skynetClusterd) Query(address any) service.Service {
	if addr, ok := address.(string); ok {
		if svc, ok := c.namedServices[addr]; ok {
			return svc
		}
	}
	return nil
}

// TODO: int address type
func (c *skynetClusterd) Register(address any, svc service.Service) error {
	if addr, ok := address.(string); ok {
		if _, ok := c.namedServices[addr]; ok {
			return fmt.Errorf("service already registered: %s", addr)
		}
		c.namedServices[addr] = svc
		return nil
	}
	return fmt.Errorf("invalid address type: %T", address)
}

func (c *skynetClusterd) Reload(config ClusterConfig) {
	if c.config == nil {
		c.config = config
		return
	}
	//self address change check
	for name, gate := range c.gate {
		if config.NodeInfo(name) != gate.Address() {
			gate.Stop()
			delete(c.gate, name)
			c.Open(name)
		}
	}
	// node address change check
	c.nodeSender.Range(func(key, value interface{}) bool {
		name := key.(string)
		client := value.(Sender)
		if config.NodeInfo(name) != client.RemoteAddr() {
			client.Exit()
			c.nodeSender.Delete(name)
		}
		return true
	})
	c.config = config
}

var mutex sync.Mutex

func (c *skynetClusterd) fetchSender(name string) Sender {
	addr := c.config.NodeInfo(name)
	if addr == "" {
		return nil
	}
	// fetch first
	if client, ok := c.nodeSender.Load(name); ok {
		return client.(Sender)
	}
	c.Lock()
	defer c.Unlock()
	// fetch again
	if client, ok := c.nodeSender.Load(name); ok {
		return client.(Sender)
	}
	client, error := NewClusterClient(c, name, addr)
	if error != nil {
		return nil
	}
	client.Start()
	c.nodeSender.Store(name, client)
	return client
}

// OnSenderExit implements Clusterd.
func (c *skynetClusterd) OnSenderExit(name string) {
	log.Printf("client removed: %s", name)
	c.nodeSender.Delete(name)
}

func (c *skynetClusterd) OnConnect(gate gate.Gate, conn net.Conn) {
	agent := NewClusterAgent(conn, c)
	agent.Start()
}

func (c *skynetClusterd) Open(name string) error {
	if c.config == nil {
		return fmt.Errorf("cluster config is nil")
	}
	addr := c.config.NodeInfo(name)
	if addr == "" {
		return fmt.Errorf("no address for node: %s", name)
	}
	c.gate[name] = gate.NewGate(
		gate.WithAddress(addr),
		gate.WithAgent(c),
	)
	return c.gate[name].Start()
}
