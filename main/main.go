package main

import (
	"fmt"

	"github.com/chtc/gmos-client/client"
)

func main() {
	cl := client.GlideinManagerClient{
		ManagerUrl: "http://gm-file-server:80",
		HostName:   "test-client",
		Port:       8080,
	}
	status, err := cl.ClientStatus()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", status)

	capability, err := cl.DoHandshake()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", capability)
}
