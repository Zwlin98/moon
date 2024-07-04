package gate

import (
	"log/slog"
	"net"
	"sync/atomic"
)

type Gate interface {
	Start() error
	Stop()
	Address() string

	AddClient()
	RemoveClient()
}

type GateOption func(*skynetGate)

type GateAgent interface {
	OnConnect(g Gate, conn net.Conn)
}

type skynetGate struct {
	address     string
	listener    net.Listener
	maxClient   int32
	clientCount int32

	agent GateAgent
}

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}

func NewGate(opt ...GateOption) *skynetGate {
	g := &skynetGate{}
	for _, o := range opt {
		o(g)
	}
	if g.maxClient == 0 {
		g.maxClient = 1024
	}
	assert(g.address != "", "Failed to create new gate, Address is empty")
	assert(g.clientCount == 0, "Failed to create new gate, ClientCount is not zero")
	assert(g.agent != nil, "Failed to create new gate, Agent is nil")
	return g
}

func (g *skynetGate) Start() (err error) {
	g.listener, err = net.Listen("tcp", g.address)
	if err != nil {
		return err
	}
	slog.Info("gate started", "address", g.address)
	go g.listenLoop()
	return nil
}

func (g *skynetGate) listenLoop() {
	for {
		conn, err := g.listener.Accept()
		if err != nil {
			slog.Error("failed to accept new client", "error", err.Error())
			continue
		}
		if g.clientCount >= g.maxClient {
			slog.Warn("client count exceed max client", "clientCount", g.clientCount, "maxClient", g.maxClient)
		}
		g.AddClient()
		slog.Info("new client connected", "remoteAddr", conn.RemoteAddr().String(), "clientCount", g.clientCount)
		g.agent.OnConnect(g, conn)
	}
}

func (g *skynetGate) Address() string {
	return g.address
}

func (g *skynetGate) Stop() {
	g.listener.Close()
}

func (g *skynetGate) AddClient() {
	atomic.AddInt32(&g.clientCount, 1)
}

func (g *skynetGate) RemoveClient() {
	atomic.AddInt32(&g.clientCount, -1)
}

func WithAddress(address string) GateOption {
	return func(g *skynetGate) {
		g.address = address
	}
}

func WithMaxClient(maxClient int32) GateOption {
	return func(g *skynetGate) {
		g.maxClient = maxClient
	}
}

func WithAgent(agent GateAgent) GateOption {
	return func(g *skynetGate) {
		g.agent = agent
	}
}
