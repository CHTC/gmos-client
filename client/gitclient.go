package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (gm *GlideinManagerClient) RepoStatus() (RepoState, error) {
	var listing RepoState
	resp, err := http.Get(gm.RouteFor("/api/public/repo-status"))
	if err != nil {
		return RepoState{}, err
	}
	defer resp.Body.Close()

	return listing, UnmarshalBody(resp.Body, &listing)
}

func (gm *GlideinManagerClient) ReportGitUsage(hash string) error {
	usage := RepoState{CommitHash: hash}

	client := &http.Client{}
	usageStr, err := json.Marshal(usage)
	req, err2 := http.NewRequest("POST", gm.RouteFor("/api/private/log-repo-access"), bytes.NewBuffer(usageStr))
	if err := errors.Join(err, err2); err != nil {
		return err
	}
	req.SetBasicAuth(gm.HostName, gm.Credentials)

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("report git usage returned status %v", resp.StatusCode)
	}
	return nil
}

func (gm *GlideinManagerClient) CloneRepo() error {

	if gm.Credentials == "" {
		return errors.New("unauthenticated client")
	}

	repo_info, err := gm.RepoStatus()
	if err != nil {
		return err
	}

	clone_dir := fmt.Sprintf("%v/%v", gm.WorkDir, repo_info.Name)
	if _, err := git.PlainClone(clone_dir, false, &git.CloneOptions{
		URL: gm.RouteFor(fmt.Sprintf("/git/%v", repo_info.Name)),
		Auth: &githttp.BasicAuth{
			Username: gm.HostName,
			Password: gm.Credentials,
		},
	}); err != nil {
		return err
	}

	return gm.ReportGitUsage(repo_info.CommitHash)
}
