package cluster

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Zwlin98/moon/gate"
	"github.com/Zwlin98/moon/lua"
)

// execute request from other skynet node
type ClusterAgent interface {
	Start()
	Exit()
}

type skynetClusterAgent struct {
	conn     net.Conn
	clusterd Clusterd
	gate     gate.Gate

	pendingReqs map[uint32]Request

	respChan chan PackedResponse
	exit     chan struct{}
}

func NewClusterAgent(gate gate.Gate, conn net.Conn, clusterd Clusterd) ClusterAgent {
	return &skynetClusterAgent{
		conn:     conn,
		clusterd: clusterd,
		gate:     gate,

		pendingReqs: make(map[uint32]Request),

		respChan: make(chan PackedResponse),
		exit:     make(chan struct{}),
	}
}

func (ca *skynetClusterAgent) safeSend(resp PackedResponse) bool {
	select {
	case <-ca.exit:
		slog.Info("ClusterAgent exited", "addr", ca.conn.RemoteAddr())
		return false
	case ca.respChan <- resp:
		return true
	}
}

func (ca *skynetClusterAgent) Start() {
	slog.Info("ClusterAgent connected", "addr", ca.conn.RemoteAddr())

	proto := gate.NewGateProto(ca.conn, ca.conn)

	// Read msg from client
	go func() {
		for {
			msg, err := proto.Read()
			if err != nil {
				slog.Error("ClusterAgent read error", "addr", ca.conn.RemoteAddr(), "error", err)
				ca.Exit()
				return
			}
			ca.dispatch(msg)
		}
	}()

	// Write response to client
	go func() {
		for {
			select {
			case <-ca.exit:
				slog.Info("ClusterAgent response channel closed", "addr", ca.conn.RemoteAddr())
				return
			case packedResp := <-ca.respChan:
				proto.Write(packedResp.Data)
				proto.WriteBatch(packedResp.Multi)
			}
		}
	}()

}

func (ca *skynetClusterAgent) Exit() {
	slog.Info("ClusterAgent exit", "addr", ca.conn.RemoteAddr())
	close(ca.exit)
	ca.gate.RemoveClient()
	(ca.conn).Close()
}

func (ca *skynetClusterAgent) dispatch(msg []byte) {
	req, err := UnpackRequest(msg)
	if err != nil {
		slog.Error("ClusterAgent dispatch error", "error", err)
	}
	session := req.Session

	if pr, ok := ca.pendingReqs[session]; ok {
		pr.Msg = append(pr.Msg, req.Msg...)
		if req.Completed {
			go ca.execute(pr)
			delete(ca.pendingReqs, session)
		} else {
			ca.pendingReqs[session] = pr
		}
	} else {
		if req.Completed {
			go ca.execute(req)
		} else {
			ca.pendingReqs[session] = req
		}
	}
}

func (ca *skynetClusterAgent) execute(req Request) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("ClusterAgent execute panic", "addr", ca.conn.RemoteAddr(), "error", r)
			ca.sendError(req, fmt.Errorf("panic: %v", r))
		}
	}()

	svc := ca.clusterd.Query(req.Address)
	if svc == nil {
		ca.sendError(req, fmt.Errorf("service not found: %v", req.Address))
		return
	}

	args, err := lua.Deserialize(req.Msg)
	if err != nil {
		ca.sendError(req, err)
		return
	}

	ret, err := svc.Execute(args)
	if err != nil {
		ca.sendError(req, err)
		return
	}

	if !req.IsPush {
		serialized, err := lua.Serialize(ret)

		if err != nil {
			ca.sendError(req, err)
		}

		resp := Response{
			Ok:      true,
			Session: req.Session,
			Msg:     serialized,
		}

		packedResp, err := PackResponse(resp)

		if err != nil {
			ca.sendError(req, err)
		}

		ok := ca.safeSend(packedResp)
		if !ok {
			slog.Error("ClusterAgent send response failed", "addr", ca.conn.RemoteAddr())
		}
	}
}

func (ca *skynetClusterAgent) sendError(req Request, err error) {
	slog.Warn("ClusterAgent send error", "addr", ca.conn.RemoteAddr(), "error", err)
	if req.IsPush {
		return
	}
	ret := []lua.Value{lua.String(err.Error())}
	serialized, _ := lua.Serialize(ret)
	resp := Response{
		Ok:      false,
		Session: req.Session,
		Msg:     serialized,
	}
	packedResp, _ := PackResponse(resp)

	ca.safeSend(packedResp)
}
