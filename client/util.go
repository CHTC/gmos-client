package client

import (
	"encoding/json"
	"fmt"
	"io"
)

// Unmarshal the body of a request or response if
// the body is readable
func UnmarshalBody[T any](bodyIo io.ReadCloser, s *T) error {
	body, err := io.ReadAll(bodyIo)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, s)
}

// Return the given request path concatenated to the top level domain
// of the client's Glidein Manager
func (gm *GlideinManagerClient) RouteFor(path string) string {
	return fmt.Sprintf("%v%v", gm.ManagerUrl, path)
}
