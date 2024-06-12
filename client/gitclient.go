package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Query the Glidein Manager for the "head" commit of the repository
func (gm *GlideinManagerClient) RepoStatus() (RepoState, error) {
	var listing RepoState
	resp, err := http.Get(gm.RouteFor("/api/public/repo-status"))
	if err != nil {
		return RepoState{}, err
	}
	defer resp.Body.Close()

	return listing, UnmarshalBody(resp.Body, &listing)
}

// Send a telemetry message to the Glidein Manager indicating successful
// checkout of a git commit
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

// Clone the repository from the upstream Glidein Manager repo
func (gm *GlideinManagerClient) cloneRepo(repo_info RepoState) error {

	clone_dir := fmt.Sprintf("%v/%v", gm.WorkDir, repo_info.Name)
	_, err := git.PlainClone(clone_dir, false, &git.CloneOptions{
		URL: gm.RouteFor(fmt.Sprintf("/git/%v", repo_info.Name)),
		Auth: &githttp.BasicAuth{
			Username: gm.HostName,
			Password: gm.Credentials,
		},
	})
	return err
}

// Reset the local copy of the git repo to the hash specified by the
// Glidein Manager's API response
func (gm *GlideinManagerClient) resetToCommit(repo_info RepoState) error {
	clone_dir := fmt.Sprintf("%v/%v", gm.WorkDir, repo_info.Name)
	repo, err1 := git.PlainOpen(clone_dir)
	worktree, err2 := repo.Worktree()
	if err := errors.Join(err1, err2); err != nil {
		return err
	}

	// git fetch
	if err := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: gm.HostName,
			Password: gm.Credentials,
		},
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	// git reset --hard <commit>
	return worktree.Reset(&git.ResetOptions{
		Commit: plumbing.NewHash(repo_info.CommitHash),
		Mode:   git.HardReset,
	})
}

// Sync the local copy repo to the state specified by the Glidein Manager
// Clone the repo if it doesn't exist locally, then hard-reset it to
// the commit reported by the Glidein Manager
func (gm *GlideinManagerClient) SyncRepo() error {
	// Check that we're authorized
	if gm.Credentials == "" {
		return errors.New("unauthenticated client")
	}

	//
	repo_info, statErr := gm.RepoStatus()
	if statErr != nil {
		return statErr
	}
	repo_dir := fmt.Sprintf("%v/%v", gm.WorkDir, repo_info.Name)

	// Clone the repo if it doesn't exist locally
	_, statErr = os.Stat(repo_dir)
	var cloneErr error
	if os.IsNotExist(statErr) {
		cloneErr = gm.cloneRepo(repo_info)
	} else if statErr != nil {
		return statErr
	}
	if cloneErr != nil {
		return cloneErr
	}

	// hard reset the local copy to the commit specified by the Glidein Manger
	if err := gm.resetToCommit(repo_info); err != nil {
		return err
	}
	return gm.ReportGitUsage(repo_info.CommitHash)
}
