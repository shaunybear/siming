package mac

import (
	"fmt"
	"os"
	"os/exec"
)

// ProcessMac MAC is running in a different process
type ProcessMac struct {
	executable string
	cmd        *exec.Cmd
	macBackend
}

// NewProcessMac Return MAC Instance
func NewProcessMac(deveui string, executable string) (mac *ProcessMac, err error) {
	backend, err := newMacBackend(deveui)

	mac = &ProcessMac{executable: executable, macBackend: backend}

	return mac, err
}

// Start the MAC
func (mac *ProcessMac) Start() (err error) {
	mac.cmd = exec.Command(mac.executable, "--deveui", mac.deveui)
	mac.cmd.Env = append(os.Environ(),
		fmt.Sprintf("MAC_RPC_BACKEND_ADDRESS=%s", rpcBackEnd))

	rpc.AddBackend(mac)
	return mac.cmd.Start()
}

// Stop the MAC
func (mac ProcessMac) Stop() {
	fmt.Printf("ProcessMac Stop not implemented\n")
}

// Command Send MAC command and return the response
func (mac *ProcessMac) Command(cmd string) (response string, err error) {
	response, err = mac.rpc.Send(mac.deveui, cmd)
	return response, err
}
