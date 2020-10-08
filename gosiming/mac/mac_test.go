package mac

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func TestMain(m *testing.M) {
	// Start MAC Services in another thread
	err := Run()
	if err != nil {
		log.Fatal(err)
	}

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestLoRaMacNode(t *testing.T) {

	deveuis := [...]string{"0000000000000004", "0000000000000005", "0000000000000006"}

	for _, deveui := range deveuis {
		mac := AddLoRaMacNode(deveui)
		mac.Start()
	}

	// Wait for macs to signal ready state
	time.Sleep(time.Duration(10*len(deveuis)) * time.Millisecond)

	// Check macs are ready
	for _, deveui := range deveuis {
		mac := Get(deveui)
		if !mac.Ready() {
			t.Errorf("Ready not received for %s", deveui)
		}
	}
}

func TestInProcMac(t *testing.T) {

	deveuis := [...]string{"0000000000000001", "0000000000000002", "0000000000000003"}

	var testMAC InProcMacFunc = func(deveui string, endpoint string) {
		sock, _ := zmq.NewSocket(zmq.REQ)
		defer sock.Close()

		sock.SetIdentity(deveui)
		if err := sock.Connect(endpoint); err != nil {
			log.Fatal(err)
		}

		//  Tell broker we're ready for work
		_, err := sock.SendMessage(ServiceReady)
		if err != nil {
			log.Fatal(err)
		}

		for {

			//  Read and save all frames until we get an empty frame
			//  In this example there is only 0 but it could be more
			msg, err := sock.RecvMessage(0)
			if err != nil {
				log.Fatal(err)
			}

			identity, msg := unwrap(msg)
			fmt.Printf("[%s] request %s\n", deveui, msg)

			sock.Send(identity, zmq.SNDMORE)
			sock.Send("", zmq.SNDMORE)
			sock.Send("OK", -1)
		}
	}

	// Start the macs
	for _, deveui := range deveuis {
		mac := AddInProcMac(deveui, testMAC)
		mac.Start()
	}

	// Wait for macs to signal ready state
	time.Sleep(time.Duration(10*len(deveuis)) * time.Millisecond)

	// Check macs are ready
	for _, deveui := range deveuis {
		mac := Get(deveui)
		if !mac.Ready() {
			t.Errorf("Ready not received for %s", deveui)
		}
	}

}
