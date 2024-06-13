package client

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// Unmarshal the body of a request or response if
// the body is readable
func UnmarshalBody[T any](bodyIo io.ReadCloser, s *T) error {
	body, err := io.ReadAll(bodyIo)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}
	return errors.Wrap(json.Unmarshal(body, s), "failed to unmarshal response body")
}

// Return the given request path concatenated to the top level domain
// of the client's Glidein Manager
func (gm *GlideinManagerClient) RouteFor(path string) string {
	return fmt.Sprintf("%v%v", gm.ManagerUrl, path)
}
