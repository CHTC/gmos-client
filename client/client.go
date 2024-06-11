package client

import (
	"net/http"
)

type GlideinManagerCredentials struct {
	Capability string
	Expires    string
}
type GlideinManagerClient struct {
	HostName string
	Port     int32

	// The URL of
	ManagerUrl string

	// TODO we might not want to store this in memory
	Credentials string
}

func (gm *GlideinManagerClient) RepoStatus() (RepoListing, error) {
	var listing RepoListing
	resp, err1 := http.Get(gm.RouteFor("/api/public/repo-status"))
	if err1 != nil {
		return RepoListing{}, err1
	}
	defer resp.Body.Close()

	return listing, UnmarshalBody(resp.Body, &listing)
}

func (gm *GlideinManagerClient) ClientStatus() ([]ClientStatus, error) {
	var statuses []ClientStatus
	resp, err1 := http.Get(gm.RouteFor("/api/public/client-status"))
	if err1 != nil {
		return []ClientStatus{}, err1
	}
	defer resp.Body.Close()

	return statuses, UnmarshalBody(resp.Body, &statuses)
}
