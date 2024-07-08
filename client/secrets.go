package client

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Get a secret value from the Glidein Manager by name
func (gm *GlideinManagerClient) ListSecrets() ([]SecretSource, error) {
	client := &http.Client{}
	secretList := []SecretSource{}

	req, err := http.NewRequest("GET", gm.RouteFor("/api/private/secrets"), nil)
	if err != nil {
		return secretList, errors.Wrap(err, "failed to create list-secrets request")
	}
	req.SetBasicAuth(gm.HostName, gm.Credentials.Capability)

	resp, err := client.Do(req)
	if err != nil {
		return secretList, errors.Wrap(err, "failed to submit list-secrets request")
	}
	if resp.StatusCode != 200 {
		return secretList, fmt.Errorf("list-secrets returned status %v", resp.StatusCode)
	}

	return secretList, UnmarshalBody(resp.Body, &secretList)
}

// Get a secret value from the Glidein Manager by name
func (gm *GlideinManagerClient) GetSecret(secretName string) (SecretValue, error) {
	client := &http.Client{}
	secretVal := SecretValue{}

	req, err := http.NewRequest("GET", gm.RouteFor(fmt.Sprintf("/api/private/secrets/%v", secretName)), nil)
	if err != nil {
		return secretVal, errors.Wrap(err, "failed to create get-secret request")
	}
	req.SetBasicAuth(gm.HostName, gm.Credentials.Capability)

	resp, err := client.Do(req)
	if err != nil {
		return secretVal, errors.Wrap(err, "failed to submit get-secret request")
	}
	if resp.StatusCode != 200 {
		return secretVal, fmt.Errorf("get-secret returned status %v", resp.StatusCode)
	}

	return secretVal, UnmarshalBody(resp.Body, &secretVal)
}
