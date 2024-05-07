package cluster

import (
	"fmt"
	"log"
	"moon/gate"
	"moon/lua"
	"net"
)

type ClusterAgent interface {
	Start()
	Exit()
}

type skynetClusterAgent struct {
	conn     net.Conn
	clusterd Clusterd

	pendingReqs map[uint32]Request

	respChan chan Response
	exit     chan struct{}
}

func NewClusterAgent(conn net.Conn, clusterd Clusterd) ClusterAgent {
	return &skynetClusterAgent{
		conn:     conn,
		clusterd: clusterd,

		pendingReqs: make(map[uint32]Request),

		respChan: make(chan Response),
		exit:     make(chan struct{}),
	}
}

func (ca *skynetClusterAgent) safeSend(resp Response) bool {
	select {
	case <-ca.exit:
		log.Printf("ClusterAgent %s exited", ca.conn.RemoteAddr())
		return false
	case ca.respChan <- resp:
		return true
	}
}

func (ca *skynetClusterAgent) Start() {
	log.Printf("ClusterAgent start from %v", ca.conn.RemoteAddr())

	proto := gate.NewGateProto(ca.conn, ca.conn)

	// Read msg from client
	go func() {
		for {
			msg, err := proto.ReadMsg()
			if err != nil {
				log.Printf("ClusterAgent %s read error: %s", ca.conn.RemoteAddr(), err)
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
				log.Printf("ClusterAgent %s response channel closed", ca.conn.RemoteAddr())
				return
			case resp := <-ca.respChan:
				packedResponse, _ := PackResponse(resp)
				proto.WriteMsg(packedResponse.Data)
				for _, msg := range packedResponse.Multi {
					proto.WriteMsg(msg)
				}
			}
		}
	}()

}

func (ca *skynetClusterAgent) Exit() {
	log.Printf("ClusterAgent %s exit", ca.conn.RemoteAddr())
	close(ca.exit)
	(ca.conn).Close()
}

func (ca *skynetClusterAgent) dispatch(msg []byte) {
	req, err := UnpackRequest(msg)
	if err != nil {
		log.Printf("ClusterAgent dispatch error: %s", err)
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
			log.Printf("ClusterAgent %s execute panic: %v", ca.conn.RemoteAddr(), r)
			ca.sendError(req, fmt.Errorf("panic: %v", r))
		}
	}()
	svc := ca.clusterd.Query(req.Address)
	if svc == nil {
		ca.sendError(req, fmt.Errorf("service not found: %v", req.Address))
		return
	}
	Args, err := lua.Deserialize(req.Msg)
	if err != nil {
		ca.sendError(req, err)
		return
	}
	ret, err := svc.Execute(Args)
	if err != nil {
		ca.sendError(req, err)
		return
	}
	if !req.IsPush {
		packedRet, err := lua.Serialize(ret)

		if err != nil {
			ca.sendError(req, err)
		}
		resp := Response{
			Ok:      true,
			Session: req.Session,
			Msg:     packedRet,
		}
		ok := ca.safeSend(resp)
		if !ok {
			log.Printf("ClusterAgent %s send response failed", ca.conn.RemoteAddr())
		}
	}
}

func (ca *skynetClusterAgent) sendError(req Request, err error) {
	log.Printf("ClusterAgent %s send error: %s", ca.conn.RemoteAddr(), err)
	ret := []lua.Value{lua.String(err.Error())}
	packed, _ := lua.Serialize(ret)
	resp := Response{
		Ok:      false,
		Session: req.Session,
		Msg:     packed,
	}
	ca.safeSend(resp)
}
