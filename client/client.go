package client

import (
	"net/http"
)

type GlideinManagerCredentials struct {
	Capability string
	Expires    string
}
type GlideinManagerClient struct {
	// The unique name of the client within its namespace
	// The Glidein manager must be separately configured to
	// Allow-list each client by name
	HostName string

	// The hostname of the Glidein Manager to connect to
	ManagerUrl string

	// The active authentication token for the client
	// TODO we might not want to store this in memory
	Credentials string

	// The base directory into which to clone repositories
	WorkDir string
}

func (gm *GlideinManagerClient) ClientStatus() ([]ClientStatus, error) {
	var statuses []ClientStatus
	resp, err := http.Get(gm.RouteFor("/api/public/client-status"))
	if err != nil {
		return []ClientStatus{}, err
	}
	defer resp.Body.Close()

	return statuses, UnmarshalBody(resp.Body, &statuses)
}
