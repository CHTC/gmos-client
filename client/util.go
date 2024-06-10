package client

import (
	"encoding/json"
	"io"
	"net/http"
)

func UnmarshalResponse[T any](resp *http.Response, s *T) error {
	body, err1 := io.ReadAll(resp.Body)
	if err1 != nil {
		return err1
	}
	return json.Unmarshal(body, s)
}
