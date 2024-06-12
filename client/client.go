package client

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
	Credentials GlideinManagerCredentials

	// The base directory into which to clone repositories
	WorkDir string
}
