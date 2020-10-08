package mac

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
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
		_, err := StartLoRaMac(deveui)
		if err != nil {
			t.Errorf("StartLoRaMac %s error %v", deveui, err)
		}
	}

	// Wait for macs to signal ready state
	time.Sleep(time.Duration(10*len(deveuis)) * time.Millisecond)

	// Check macs are ready
	for _, deveui := range deveuis {
		mac := Get(deveui)
		if !mac.IsConnected() {
			t.Errorf("Ready not received for %s", deveui)
		}
	}
}

func testMacWorker(deveui string, count int, wg *sync.WaitGroup) {
	defer wg.Done()

	mac := Get(deveui)

	for i := 0; i < count; i++ {
		time.Sleep(100 * time.Millisecond)
		request := "Hello " + deveui
		fmt.Printf("Worker %s request #%d: %s\n", deveui, i, request)
		reply, err := mac.Request(request)
		if err == nil {
			fmt.Printf("Worker %s reply   #%d: %s\n", deveui, i, reply)
		} else {
			fmt.Printf("Worker %s reply error %v\n", deveui, err)
		}
	}
	fmt.Printf("Worker %s done\n", deveui)
}

func TestInProcMac(t *testing.T) {
	var wg sync.WaitGroup
	var deveuis []string
	var devEuiCount uint64 = 500
	requestCount := 100

	var i uint64
	for i = 1; i <= devEuiCount; i++ {
		deveui := strconv.FormatUint(i, 16)
		deveuis = append(deveuis, deveui)
	}

	for _, deveui := range deveuis {
		_, err := StartInProcMac(deveui, testMAC)
		if err != nil {
			t.Errorf("StartInProcMac %s error %v", deveui, err)
		}
	}

	time.Sleep(5 * time.Second)

	for _, deveui := range deveuis {
		wg.Add(1)
		go testMacWorker(deveui, requestCount, &wg)
	}

	wg.Wait()
	fmt.Println("DONE Waiting")
}

var testMAC InProcMacFunc = func(deveui string, endpoint string) {
	sock, _ := zmq.NewSocket(zmq.REQ)
	defer sock.Close()

	mac := macBackend{deveui: deveui}

	sock.SetIdentity(mac.Identity())

	if err := sock.Connect(endpoint); err != nil {
		log.Fatal(err)
	}

	//  Tell broker we're ready for work
	_, err := sock.SendMessage(BackendReady)
	if err != nil {
		log.Fatal(err)
	}

	for {

		//  Read and save all frames until we get an empty frame
		//  In this example there is only 0 but it could be more
		identity, _ := sock.Recv(0)
		empty, _ := sock.Recv(0)
		if empty != "" {
			panic(fmt.Sprintf("empty is not \"\":%q", empty))

		}

		request, _ := sock.Recv(0)

		sock.Send(identity, zmq.SNDMORE)
		sock.Send("", zmq.SNDMORE)
		sock.Send(fmt.Sprintf("%s from %s\n", request, deveui), 0)
	}
}
