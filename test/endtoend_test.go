package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chtc/gmos-client/client"
)

func TestEndToEnd(t *testing.T) {
	cl := client.GlideinManagerClient{
		ManagerUrl: "http://gm-file-server:80",
		HostName:   "test-client",
		WorkDir:    ".",
	}
	// Ensure handshake is idempotent
	for i := 0; i < 2; i++ {
		if err := cl.DoHandshake(8080); err != nil {
			t.Fatalf("Handshake failed: %v", err)
		}
		fmt.Printf("%+v\n", cl.Credentials)
	}

	// Ensure SyncRepo is idempotent
	update, err := cl.SyncRepo()
	if err != nil {
		t.Fatalf("Sync Repo #1 failed: %v", err)
	}
	assert.True(t, update.Created())
	assert.True(t, update.Updated())

	update, err = cl.SyncRepo()
	if err != nil {
		t.Fatalf("Sync Repo #2 failed: %v", err)
	}
	assert.False(t, update.Created())
	assert.False(t, update.Updated())

}
