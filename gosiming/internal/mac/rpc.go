//
//  Siming Mac layer RPC framework
//

package mac

import (
	"fmt"
	"sync"
	"time"

	zmq "github.com/pebbe/zmq4"
)

// Constants
const (
	BackendReady = "\001" //  Signals service is ready
)

// RPC  Sends requests to backend service and returns responses to requestor
type RPC struct {
	frontend  *zmq.Socket //  Listen to clients
	backend   *zmq.Socket //  Listen to services
	reactor   *zmq.Reactor
	backends  map[string]Backend
	bendpoint string
	fendpoint string
	mux       sync.Mutex
}

// Backend defines the interface used to track backend service
type Backend interface {
	Identity() string
	SetConnectedState(status bool)
	IsConnected() bool
}

// RPCRequest RPC Requester
type RPCRequest struct {
	sock     *zmq.Socket
	frontend string
}

// FrontEnd RPC transport name
func (request RPCRequest) FrontEnd() string {
	return request.frontend
}

// NewRPC Creates a New RPC Broker
func NewRPC(frontend string, backend string) (rpc *RPC, err error) {
	fsock, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return nil, err
	}

	err = fsock.Bind(frontend)
	if err != nil {
		return nil, err
	}

	bsock, err := zmq.NewSocket(zmq.ROUTER)
	err = bsock.Bind(backend)
	if err != nil {
		return nil, err
	}

	b := &RPC{frontend: fsock,
		backend:   bsock,
		reactor:   zmq.NewReactor(),
		fendpoint: frontend,
		bendpoint: backend,
		backends:  make(map[string]Backend)}

	return b, nil
}

//  In the reactor design, each time a message arrives on a socket, the
//  reactor passes it to a handler function. We have two handlers; one
//  for the frontend, one for the backend:

//handleFrontEnd Handles input from frontend
func handleFrontend(rpc *RPC) error {
	//  Get client request, routed with identity added by client
	msg, err := rpc.frontend.RecvMessage(0)
	if err != nil {
		return err
	}

	client, msg := unwrap(msg)
	backend, msg := unwrap(msg)

	rpc.backend.Send(backend, zmq.SNDMORE)
	rpc.backend.Send("", zmq.SNDMORE)
	rpc.backend.Send(client, zmq.SNDMORE)
	rpc.backend.Send("", zmq.SNDMORE)
	rpc.backend.Send(msg[0], 0)

	return nil
}

// handleBackend Handles input from backend
// func handleBackend(fsock *zmq.Socket, bsock *zmq.Socket) error {
func handleBackend(rpc *RPC) error {

	msg, err := rpc.backend.RecvMessage(0)
	if err != nil {
		fmt.Printf("[BACKEND] RecvMessage error %v\n", err)
		return err
	}
	backend, msg := unwrap(msg)

	//  Forward message to client if it's not a READY
	if msg[0] != BackendReady {
		rpc.frontend.SendMessage(msg)
	} else {
		rpc.mux.Lock()
		b, _ := rpc.backends[backend]
		if b != nil {
			b.SetConnectedState(true)
		} else {
			fmt.Printf("[BACKEND] received connected from unknown backend %s\n", backend)
		}
		rpc.mux.Unlock()
	}

	return nil
}

// AddBackend Adds RPC service
func (rpc *RPC) AddBackend(backend Backend) {
	backend.SetConnectedState(false)
	rpc.mux.Lock()
	rpc.backends[backend.Identity()] = backend
	rpc.mux.Unlock()
}

//Run Fires up the RPC Broker
func (rpc *RPC) Run() (err error) {
	rpc.reactor.AddSocket(rpc.backend, zmq.POLLIN,
		func(e zmq.State) error { return handleBackend(rpc) })

	rpc.reactor.AddSocket(rpc.frontend, zmq.POLLIN,
		func(e zmq.State) error { return handleFrontend(rpc) })

	go func() {
		err = rpc.reactor.Run(-1)
	}()

	// Wait for reactor to start
	timeout := time.After(2 * time.Second)
	<-timeout

	return err
}

// NewRPCRequest Creates and connects a RPC client
func (rpc *RPC) NewRPCRequest() *RPCRequest {
	sock, _ := zmq.NewSocket(zmq.REQ)
	sock.Connect(rpc.fendpoint)
	r := &RPCRequest{sock: sock, frontend: rpc.fendpoint}
	return r
}

//Send Request
func (request *RPCRequest) Send(service string, msg string) (reply string, err error) {
	// Send request
	request.sock.Send(service, zmq.SNDMORE)
	request.sock.Send("", zmq.SNDMORE)
	request.sock.Send(msg, 0)

	// Wait for reply
	return request.sock.Recv(0)
}

//  unwrap  pops frame off front of message and returns it as 'head'
//  If next frame is empty, pops that empty frame.
//  Return remaining frames of message as 'tail'
func unwrap(msg []string) (head string, tail []string) {
	head = msg[0]

	if len(msg) > 1 && msg[1] == "" {
		tail = msg[2:]
	} else {
		tail = msg[1:]
	}
	return
}
