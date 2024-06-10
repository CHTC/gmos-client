package main

import (
	"fmt"

	"github.com/chtc/gmos-client/client"
)

func main() {
	cl := client.GlideinManagerClient{ManagerUrl: "http://localhost:8080"}
	status, err := cl.ClientStatus()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", status)
}
