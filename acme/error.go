package acme

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error asdf
type Error struct {
	StatusCode int    `json:"status,omitempty"`
	Type       string `json:"type"`
	Detail     string `json:"detail"`
}

func (e Error) Error() string {
	return fmt.Sprintf("[%d] Type: %s Detail: %s", e.StatusCode, e.Type, e.Detail)
}

func handleHTTPError(resp *http.Response) error {
	var errorDetail Error
	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&errorDetail)
	if err != nil {
		return err
	}

	errorDetail.StatusCode = resp.StatusCode
	return errorDetail
}
