package test

import (
	"encoding/base64"
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

func TestSecrets(t *testing.T) {
	cl := client.GlideinManagerClient{
		ManagerUrl: "http://gm-file-server:80",
		HostName:   "test-client",
		WorkDir:    ".",
	}
	sampleSecret := "sample.secret"
	sampleSecretVal := "Hello, Secret!\n"

	// Ensure handshake is idempotent
	if err := cl.DoHandshake(8080); err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	secrets, err := cl.ListSecrets()
	if err != nil {
		t.Fatalf("Retrieve secret failed: %+v", err)
	}
	assert.Equal(t, 1, len(secrets))
	assert.Equal(t, sampleSecret, secrets[0].Name)

	secret, err := cl.GetSecret("sample.secret")
	if err != nil {
		t.Fatalf("Retrieve secret failed: %+v", err)
	}
	decodedSecret, err := base64.StdEncoding.DecodeString(secret.Value)
	if err != nil {
		t.Fatalf("Failed to decode secret value %v", err)
	}
	assert.Equal(t, sampleSecret, secret.Name)
	assert.Equal(t, sampleSecretVal, string(decodedSecret))
}
