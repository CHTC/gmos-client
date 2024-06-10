package client

import (
	"fmt"
	"net/http"
)

type GlideinManagerClient struct {
	HostName string
	Port     int32

	ManagerUrl string
}

func (gm *GlideinManagerClient) RepoStatus() (RepoListing, error) {
	var listing RepoListing
	resp, err1 := http.Get(fmt.Sprintf("%v/api/public/repo-status", gm.ManagerUrl))
	if err1 != nil {
		return RepoListing{}, err1
	}
	defer resp.Body.Close()

	return listing, UnmarshalResponse(resp, &listing)
}

func (gm *GlideinManagerClient) ClientStatus() ([]ClientStatus, error) {
	var statuses []ClientStatus
	resp, err1 := http.Get(fmt.Sprintf("%v/api/public/client-status", gm.ManagerUrl))
	if err1 != nil {
		return []ClientStatus{}, err1
	}
	defer resp.Body.Close()

	return statuses, UnmarshalResponse(resp, &statuses)
}
