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
	Ready() bool
	SetReady(isReady bool)
}

type macState struct {
	deveui string
	ready  bool
	rpc    *RPCRequest
}

func (mac *macState) SetReady(ready bool) {
	fmt.Printf("[%s] ready=%v\n", mac.deveui, ready)
	mac.ready = ready
}

func (mac *macState) DevEui() string {
	return mac.deveui
}

// Name Service interface
func (mac *macState) Name() string {
	return mac.deveui
}

func (mac *macState) Ready() bool {
	return mac.ready
}

// Command Send MAC command and return the response
func (mac *macState) Command(cmd string) (response string, err error) {
	response, err = mac.rpc.Send(mac.deveui, cmd)
	return response, err
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

// AddLoRaMacNode Add a Node configured with the LoRaMac-node stack
func AddLoRaMacNode(deveui string) Mac {
	mac := NewProcessMac(deveui, loraMacNodeExecutable)
	macs[deveui] = mac
	return mac
}

// AddInProcMac Adds an Inproc  test MAC for testing obviously
func AddInProcMac(deveui string, f InProcMacFunc) Mac {
	mac := NewInProcMac(deveui, f)
	macs[deveui] = mac
	return mac
}

// Get Returns Node instance
func Get(deveui string) Mac {
	m, ok := macs[deveui]
	if !ok {
		return nil
	}
	return m
}
