package mac

import (
	"fmt"
	"log"
	"sync"
)

const (
	rpcFrontEnd           = "inproc://mac.rpc"
	rpcBackEnd            = "ipc:///opt/siming/zmq/mac.rpc"
	loraMacNodeExecutable = "/opt/siming/bin/loRaMac-node"
)

var (
	initMacOnce sync.Once
	rpc         *RPC
	macs        map[string]Mac
)

// Mac Interface
type Mac interface {
	Start() error
	Stop()
	DevEui() string
	IsConnected() bool
	SetConnectedState(connected bool)
	Request(command string) (reply string, err error)
}

type macBackend struct {
	deveui    string
	identity  string
	connected bool
	rpc       *RPCRequest
}

// NewMacBackend Construct a MAC Backend
func newMacBackend(deveui string) (backend macBackend, err error) {

	backend = macBackend{
		deveui:    deveui,
		connected: false,
		rpc:       rpc.NewRPCRequest(),
	}

	return backend, nil
}

func (mac *macBackend) SetConnectedState(connected bool) {
	fmt.Printf("Set mac backend %s connected %v => %v\n", mac.deveui, mac.connected, connected)
	mac.connected = connected
}

func (mac macBackend) IsConnected() bool {
	return mac.connected
}

func (mac macBackend) DevEui() string {
	return mac.deveui
}

// Backend Identity
func (mac macBackend) Identity() string {
	return mac.deveui
}

// Command Send MAC command and return the response
func (mac macBackend) Request(cmd string) (reply string, err error) {
	// Check the backend is connected
	if !mac.IsConnected() {
		return "", fmt.Errorf(fmt.Sprintf("%s is not connected", mac.deveui))
	}

	return mac.rpc.Send(mac.deveui, cmd)

}

// Run Initializes and starts MAC services
func Run() error {
	var err error

	onceBody := func() {
		macs = make(map[string]Mac)
		rpc, err = NewRPC(rpcFrontEnd, rpcBackEnd)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("MAC Initializing RPC\n")
		err = rpc.Run()
	}

	initMacOnce.Do(onceBody)

	return err
}

// StartLoRaMac Add a Node configured with the LoRaMac-node stack
func StartLoRaMac(deveui string) (mac Mac, err error) {
	mac, err = NewProcessMac(deveui, loraMacNodeExecutable)
	if err == nil {
		err = mac.Start()
	}

	macs[deveui] = mac
	return mac, err
}

// StartInProcMac  Adds an Inproc  test MAC for testing obviously
func StartInProcMac(deveui string, f InProcMacFunc) (mac Mac, err error) {
	mac, err = NewInProcMac(deveui, f)

	if err == nil {
		err = mac.Start()
	}

	macs[deveui] = mac
	return mac, err
}

// Get Returns Node instance
func Get(deveui string) Mac {
	m, ok := macs[deveui]
	if !ok {
		return nil
	}
	return m
}
