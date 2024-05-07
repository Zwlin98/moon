package cluster

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/Zwlin98/moon/gate"
	"github.com/Zwlin98/moon/lua"
)

// call/send other skynet node
type Sender interface {
	RemoteAddr() string

	Call(string, string, []lua.Value) ([]lua.Value, error)
	Send(string, string, []lua.Value) error

	Start()
	Exit()
}

type skynetSender struct {
	clusterd   Clusterd
	remoteName string
	remoteAddr string
	conn       net.Conn

	session uint32

	pendingResponse map[uint32]Response
	pendingRespChan sync.Map

	reqChan chan PackedRequest
	exit    chan struct{}
}

func NewClusterClient(clusterd Clusterd, name string, addr string) (Sender, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client := skynetSender{
		clusterd:        clusterd,
		remoteName:      name,
		remoteAddr:      addr,
		conn:            conn,
		session:         1,
		reqChan:         make(chan PackedRequest),
		pendingResponse: make(map[uint32]Response),
		exit:            make(chan struct{}),
	}

	return &client, nil
}

func (sc *skynetSender) Start() {
	proto := gate.NewGateProto(sc.conn, sc.conn)

	go func() {
		for {
			msg, err := proto.ReadMsg()
			if err != nil {
				log.Printf("ClusterClient %s failed to read message, %s", sc.name(), err.Error())
				sc.Exit()
				return
			}
			sc.dispatch(msg)
		}
	}()

	go func() {
		for {
			select {
			case <-sc.exit:
				log.Printf("ClusterClient %s exited [W]", sc.name())
				return
			case req := <-sc.reqChan:
				err := proto.WriteMsg(req.Data)
				if err != nil {
					log.Printf("ClusterClient %s failed to write message, %s", sc.name(), err.Error())
					sc.Exit()
					return
				}
				for _, part := range req.Multi {
					err = proto.WriteMsg(part)
					if err != nil {
						log.Printf("ClusterClient %s failed to write message, %s", sc.name(), err.Error())
						sc.Exit()
						return
					}
				}
			}
		}
	}()
}

func (sc *skynetSender) callRet(resp Response) {
	session := resp.Session
	retChan, ok := sc.pendingRespChan.Load(session)
	if ok {
		retChan.(chan Response) <- resp
	} else {
		log.Printf("session %d, ClusterClient %s callRet failed, no pending response", session, sc.name())
	}
}

func (sc *skynetSender) callError(session uint32, msg string) {
	sc.callRet(Response{
		Session: session,
		Ok:      false,
		Msg:     []byte(msg),
	})
}

func (sc *skynetSender) packCall(service string, method string, args []lua.Value, isPush bool) (PackedRequest, uint32, error) {
	realArgs := []lua.Value{lua.String(method)}
	realArgs = append(realArgs, args...)

	packedArgs, err := lua.Serialize(realArgs)
	if err != nil {
		return PackedRequest{}, 0, err
	}

	session := atomic.AddUint32(&sc.session, 1)

	req := Request{
		Address: service,
		Session: session,
		IsPush:  isPush,
		Msg:     packedArgs,
	}

	packedReq, err := PackRequest(req)

	return packedReq, session, err
}

// Call implements Client.
func (sc *skynetSender) Call(service string, method string, args []lua.Value) ([]lua.Value, error) {
	packReq, session, err := sc.packCall(service, method, args, false)

	if err != nil {
		return nil, err
	}

	respChan := make(chan Response)
	sc.pendingRespChan.Store(session, respChan)
	defer sc.pendingRespChan.Delete(session)

	select {
	case <-sc.exit:
		return nil, fmt.Errorf("ClusterClient %s is exited [CallOut]", sc.name())
	case sc.reqChan <- packReq:
	}

	select {
	case <-sc.exit:
		return nil, fmt.Errorf("session %d, ClusterClient %s is exited [Waiting CallRet]", session, sc.name())
	case resp := <-respChan:
		if resp.Ok {
			return lua.Deserialize(resp.Msg)
		} else {
			return nil, fmt.Errorf("remote call failed: %s", resp.Msg)
		}
	}
}

func (sc *skynetSender) Send(service string, method string, args []lua.Value) error {
	packReq, _, err := sc.packCall(service, method, args, true)
	if err != nil {
		return err
	}
	select {
	case <-sc.exit:
		return fmt.Errorf("ClusterClient %s is exited [SendOut]", sc.name())
	case sc.reqChan <- packReq:
		return nil
	}
}

func (sc *skynetSender) Exit() {
	log.Printf("ClusterClient %s exit", sc.name())
	close(sc.exit)
	sc.conn.Close()
	sc.clusterd.OnSenderExit(sc.name())
}

func (sc *skynetSender) RemoteAddr() string {
	return sc.remoteAddr
}

func (sc *skynetSender) dispatch(msg []byte) {
	resp, err := UnpackResponse(msg)
	if err != nil {
		log.Printf("failed to unpack response, %s", err.Error())
		return
	}
	switch resp.Padding {
	case RESPONSE_END:
		sc.callRet(resp)
	case RESPONSE_MULTI_BEGIN:
		sc.pendingResponse[resp.Session] = resp
	case RESPONSE_MULTI_PART:
		prevResp, ok := sc.pendingResponse[resp.Session]
		if !ok {
			log.Printf("unexpected multi part response")
			sc.callError(resp.Session, "unexpected multi part response")
			return
		} else {
			prevResp.Msg = append(prevResp.Msg, resp.Msg...)
			sc.pendingResponse[resp.Session] = prevResp
		}
	case RESPONSE_MULTI_END:
		prevResp, ok := sc.pendingResponse[resp.Session]
		if !ok {
			log.Printf("unexpected multi end response")
			sc.callError(resp.Session, "unexpected multi end response")
			return
		} else {
			prevResp.Msg = append(prevResp.Msg, resp.Msg...)
			sc.callRet(prevResp)
			delete(sc.pendingResponse, resp.Session)
		}
	}
}

func (sc *skynetSender) name() string {
	return sc.remoteName
}
