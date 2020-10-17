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
	var wg sync.WaitGroup
	var deveuis []string
	var devEuiCount uint64 = 1
	requestCount := 1

	var i uint64
	for i = 1; i <= devEuiCount; i++ {
		deveui := strconv.FormatUint(i, 16)
		deveuis = append(deveuis, deveui)
	}

	for _, deveui := range deveuis {
		_, err := StartLoRaMac(deveui)
		if err != nil {
			t.Errorf("StartLoraMac %s error %v", deveui, err)
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
