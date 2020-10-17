package mac

import "fmt"

// InProcMacFunc In process MAC function type
type InProcMacFunc func(deveui string, endpoint string)

// InProcMac MAC is running in a different process
type InProcMac struct {
	macBackend
	f InProcMacFunc
}

// Start the MAC
func (mac *InProcMac) Start() (err error) {
	rpc.AddBackend(mac)
	go mac.f(mac.deveui, rpc.bendpoint)
	return nil
}

// Stop the MAC
func (mac InProcMac) Stop() {
	fmt.Printf("InProcMac Stop not implemented\n")
}

// NewInProcMac Return MAC Instance
func NewInProcMac(deveui string, f InProcMacFunc) (mac *InProcMac, err error) {
	backend, err := newMacBackend(deveui)

	mac = &InProcMac{macBackend: backend, f: f}

	return mac, err
}
