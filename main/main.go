package main

import (
	"fmt"

	"github.com/chtc/gmos-client/client"
)

func main() {
	cl := client.GlideinManagerClient{
		ManagerUrl: "http://gm-file-server:80",
		HostName:   "test-client",
	}
	status, err := cl.ClientStatus()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", status)

	if err := cl.DoHandshake(8080); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", cl.Credentials)
}
