package client

import (
	"encoding/json"
	"io"
)

func UnmarshalBody[T any](bodyIo io.ReadCloser, s *T) error {
	body, err1 := io.ReadAll(bodyIo)
	if err1 != nil {
		return err1
	}
	return json.Unmarshal(body, s)
}
