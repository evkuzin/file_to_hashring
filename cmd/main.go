package main

import (
	"file-to-hashring/src/http_server"
)

func main() {

	//ring := hashring.New(pgServers)
	//for i := 0; i < len(pgServers); i++ {
	//
	//	server := ring.GetServer(fmt.Sprintf("%s_%d", filename, i))
	//
	//}
	//for srv, _ := range stats {
	//	fmt.Printf("%s: %d\n", srv, stats[srv])
	//}
	http_server.Start()
}
