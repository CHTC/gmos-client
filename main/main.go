package main

import (
	"fmt"

	"github.com/chtc/gmos-client/client"
)

func main() {
	cl := client.GlideinManagerClient{
		ManagerUrl: "http://gm-file-server:80",
		HostName:   "test-client",
		WorkDir:    ".",
	}
	if err := cl.DoHandshake(8080); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", cl.Credentials)

	for i := 0; i < 2; i++ {
		if err := cl.SyncRepo(); err != nil {
			panic(err)
		}
	}
}
