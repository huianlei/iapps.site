// network.go
package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"iapps.site/network"
)

// go run app.go -connect localhost:9001
func main() {
	connect := flag.String("connect", "", "IP address of process to join. If empty, go into listen mode.")
	count := flag.Int("count", 1, "client count")
	concurrent := flag.Int("concurrent", 1, "client concurrent num in one second")
	flag.Parse()

	// If the connect flag is set, go into client mode.
	if *connect != "" {
		client(*connect, *count, *concurrent)
		log.Println("Client done.")
		return
	}

	err := network.BootServer()
	if err != nil {
		log.Printf("Error:%#v", err)
	}

	log.Println("Server done.")
}

var client_latch sync.WaitGroup

func client(addr string, clientCount int, concurrent int) {
	// init
	network.ActionType = network.ActionTypeForClient
	network.GetActionManagerIns()

	var sleepNano int64 = time.Second.Nanoseconds() / int64(concurrent)
	log.Printf("Client boot param clientCount=%d, concurrent=%d, sleepNano=%d", clientCount, concurrent, sleepNano)

	var client_id uint64 = uint64(0)
	for i := 1; i <= clientCount; i++ {
		client_latch.Add(1)
		go func() {
			//log.Printf("Try New TcpClient_%d\n", i)
			cid := atomic.AddUint64(&client_id, 1)
			client, err := network.NewTcpClient(addr, cid)
			if err != nil {
				log.Printf("Create TcpClient error%#v\n", err)
				client_latch.Done()
				runtime.Goexit()
				return
			}
			if err == nil {
				client.InitAccount(fmt.Sprintf("acc_%d", cid), fmt.Sprintf("robot_%d", cid))
				client.Start()
			}
		}()

		time.Sleep(time.Duration(sleepNano * time.Nanosecond.Nanoseconds()))
	}

	client_latch.Wait()
}
