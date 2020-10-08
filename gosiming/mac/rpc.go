//
//  Siming Mac layer RPC framework
//

package mac

import (
	"fmt"
	"time"

	zmq "github.com/pebbe/zmq4"
)

// Constants
const (
	ServiceReady = "\001" //  Signals service is ready
)

// RPC  Sends requests to backend service and returns responses to requestor
type RPC struct {
	frontend  *zmq.Socket //  Listen to clients
	backend   *zmq.Socket //  Listen to services
	reactor   *zmq.Reactor
	services  map[string]Service
	bendpoint string
	fendpoint string
}

// Service defines the service interface used to track service state
type Service interface {
	Name() string
	SetReady(status bool)
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

	fmt.Printf("RPC bind frontend: %s\n", frontend)
	err = fsock.Bind(frontend)
	if err != nil {
		fmt.Printf("RPC bind frontend %s error %s\n", frontend, err)
		return nil, err
	}

	fmt.Printf("RPC bind backend: %s\n", backend)
	bsock, err := zmq.NewSocket(zmq.ROUTER)
	err = bsock.Bind(backend)
	if err != nil {
		fmt.Printf("RPC bind backend %s error %s\n", backend, err)
		return nil, err
	}

	b := &RPC{frontend: fsock,
		backend:   bsock,
		reactor:   zmq.NewReactor(),
		fendpoint: frontend,
		bendpoint: backend,
		services:  make(map[string]Service)}

	return b, nil
}

//  In the reactor design, each time a message arrives on a socket, the
//  reactor passes it to a handler function. We have two handlers; one
//  for the frontend, one for the backend:

//handleFrontEnd Handles input from frontend
// func handleFrontend(fsock *zmq.Socket, bsock *zmq.Socket) error {
func handleFrontend(rpc *RPC) error {
	//  Get client request, routed with identity added by client
	msg, err := rpc.frontend.RecvMessage(0)
	// msg, err := fsock.RecvMessage(0)
	if err != nil {
		return err
	}

	client, msg := unwrap(msg)
	service, msg := unwrap(msg)

	fmt.Printf("[FRONTEND] request to %s\n", service)
	/*
		bsock.Send(service, zmq.SNDMORE)
		bsock.Send("", zmq.SNDMORE)
		bsock.Send(client, zmq.SNDMORE)
		bsock.Send("", zmq.SNDMORE)
		bsock.Send(msg[0], 0) */

	rpc.backend.Send(service, zmq.SNDMORE)
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
	// msg, err := bsock.RecvMessage(0)
	if err != nil {
		return err
	}
	service, msg := unwrap(msg)

	//  Forward message to client if it's not a READY
	if msg[0] != ServiceReady {
		rpc.frontend.SendMessage(msg)
	} else {
		s, _ := rpc.services[service]
		if s != nil {
			s.SetReady(true)
		} else {
			fmt.Printf("[BACKEND] received ready state from unadded service %s\n", service)
		}
	}

	return nil
}

// AddService Adds RPC service
func (rpc *RPC) AddService(service Service) {
	service.SetReady(false)
	rpc.services[service.Name()] = service
}

//Run Fires up the RPC Broker
func (rpc *RPC) Run() (err error) {
	rpc.reactor.AddSocket(rpc.backend, zmq.POLLIN,
		func(e zmq.State) error { return handleBackend(rpc) })

	rpc.reactor.AddSocket(rpc.frontend, zmq.POLLIN,
		func(e zmq.State) error { return handleFrontend(rpc) })

	// Reactor blocks so run it in a goroutine
	ready := make(chan bool)

	go func() {
		ready <- true
		err = rpc.reactor.Run(-1)
	}()

	// Wait for ready signal
	<-ready

	// Total amature move here. I don't want to return until the reactor has run, but I'm not
	// sure yet on how to get the reactor's goroutine state
	fmt.Println("Novice Go Developer move with the 1 second delay on RPC startup")
	time.Sleep(1 * time.Second)

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
	fmt.Printf("[CLIENT] request to %s\n", service)
	request.sock.Send(service, zmq.SNDMORE)
	request.sock.Send("", zmq.SNDMORE)
	request.sock.Send(msg, 0)

	// Wait for reply
	reply, err = request.sock.Recv(0)
	if err != nil {
		fmt.Printf("[CLIENT] reply from %s error %s\n", service, err)
	} else {
		fmt.Printf("[CLIENT] reply from %s = %s\n", service, reply)
	}

	return reply, err
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
