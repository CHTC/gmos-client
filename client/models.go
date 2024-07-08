package client

type ChallengeInitiateRequest struct {
	ClientName      string `json:"client_name"`
	CallbackAddress string `json:"callback_address"`
}

type ChallengeInitiateResponse struct {
	IdSecret        string `json:"id_secret"`
	ChallengeSecret string `json:"challenge_secret"`
}

type ChallengeCompleteRequest struct {
	IdSecret   string `json:"id_secret"`
	Capability string `json:"capability"`
	Expires    string `json:"expires"`
}

type ChallengeCompleteResponse struct {
	ChallengeSecret string `json:"challenge_secret"`
}

type RepoState struct {
	Name       string `json:"name"`
	Upstream   string `json:"upstream"`
	CommitHash string `json:"commit_hash"`
}

type ClientAccessStatus struct {
	AccessTime string `json:"access_time"`
	CommitHash string `json:"commit_hash"`
}

type ClientAuthState struct {
	State     string `json:"state"`
	Initiated string `json:"initiated"`
	Expires   string `json:"expires"`
}

type ClientStatus struct {
	ClientName string             `json:"client_name"`
	AuthState  ClientAuthState    `json:"auth_state"`
	RepoAccess ClientAccessStatus `json:"repo_access"`
}

type SecretSource struct {
	Name    string `json:"secret_name"`
	Version string `json:"secret_version"`
	Source  string `json:"secret_source"`
}

type SecretValue struct {
	Name    string `json:"secret_name"`
	Version string `json:"secret_version"`
	Value   string `json:"secret_value"`
}
