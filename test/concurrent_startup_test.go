package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/chtc/gmos-client/client"
)

func _TestSimultaneousStartup(t *testing.T) {
	attempts := 5
	cl := client.GlideinManagerClient{
		ManagerUrl: "http://gm-file-server:80",
		HostName:   "test-client-concurrent",
		WorkDir:    ".",
	}
	// Ensure handshake is idempotent
	for i := 0; i < attempts; i++ {
		if err := cl.DoHandshake(8080); err != nil {
			fmt.Printf("%v", err)
		} else {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("Handshake failed after %v attempts", attempts)
	fmt.Printf("%+v\n", cl.Credentials)
}
