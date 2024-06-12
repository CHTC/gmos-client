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

type RepoUpdate struct {
	PreviousCommit string
	CurrentCommit  string
}

func (ru *RepoUpdate) Created() bool {
	return ru.PreviousCommit == "" && ru.CurrentCommit != ""
}

func (ru *RepoUpdate) Updated() bool {
	return ru.PreviousCommit != ru.CurrentCommit
}

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
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", gm.RouteFor("/api/private/log-repo-access"), bytes.NewBuffer(usageStr))
	if err != nil {
		return err
	}

	req.SetBasicAuth(gm.HostName, gm.Credentials.Capability)

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
func (gm *GlideinManagerClient) cloneRepo(repoInfo RepoState) error {
	cloneDir := fmt.Sprintf("%v/%v", gm.WorkDir, repoInfo.Name)
	_, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL: gm.RouteFor(fmt.Sprintf("/git/%v", repoInfo.Name)),
		Auth: &githttp.BasicAuth{
			Username: gm.HostName,
			Password: gm.Credentials.Capability,
		},
	})
	return err
}

func getCurrentCommit(repoDir string) (string, error) {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	return head.Hash().String(), nil
}

// Reset the local copy of the git repo to the hash specified by the
// Glidein Manager's API response
func (gm *GlideinManagerClient) resetToCommit(repoInfo RepoState) error {
	cloneDir := fmt.Sprintf("%v/%v", gm.WorkDir, repoInfo.Name)
	repo, err := git.PlainOpen(cloneDir)

	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// git fetch
	if err := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: gm.HostName,
			Password: gm.Credentials.Capability,
		},
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	// git reset --hard <commit>
	return worktree.Reset(&git.ResetOptions{
		Commit: plumbing.NewHash(repoInfo.CommitHash),
		Mode:   git.HardReset,
	})
}

// Sync the local copy repo to the state specified by the Glidein Manager
// Clone the repo if it doesn't exist locally, then hard-reset it to
// the commit reported by the Glidein Manager
func (gm *GlideinManagerClient) SyncRepo() (RepoUpdate, error) {
	// Check that we're authorized
	repoUpdate := RepoUpdate{}
	if gm.Credentials == (GlideinManagerCredentials{}) {
		return repoUpdate, errors.New("unauthenticated client")
	}

	// Retrieve the desired active commit from
	repoInfo, err := gm.RepoStatus()
	if err != nil {
		return repoUpdate, err
	}
	repoUpdate.CurrentCommit = repoInfo.CommitHash

	// Clone the repo if it doesn't exist locally
	repoDir := fmt.Sprintf("%v/%v", gm.WorkDir, repoInfo.Name)
	_, statErr := os.Stat(repoDir)
	isNewRepo := os.IsNotExist(statErr)
	if statErr != nil && !isNewRepo {
		return repoUpdate, statErr
	}
	if isNewRepo {
		if err := gm.cloneRepo(repoInfo); err != nil {
			return repoUpdate, err
		}
	} else {
		repoUpdate.PreviousCommit, err = getCurrentCommit(repoDir)
		if err != nil {
			return repoUpdate, err
		}
	}

	// hard reset the local copy to the commit specified by the Glidein Manger
	if err := gm.resetToCommit(repoInfo); err != nil {
		return repoUpdate, err
	}

	// If the local checkout of the repo was updated, report telemetry back to the server
	if repoUpdate.Updated() {
		err = gm.ReportGitUsage(repoInfo.CommitHash)
	}

	return repoUpdate, err
}
