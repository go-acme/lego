package acme

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	tosAgreementError = "Must agree to subscriber agreement before any further actions"
)

// RemoteError is the base type for all errors specific to the ACME protocol.
type RemoteError struct {
	StatusCode int    `json:"status,omitempty"`
	Type       string `json:"type"`
	Detail     string `json:"detail"`
}

func (e RemoteError) Error() string {
	return fmt.Sprintf("acme: Error %d - %s - %s", e.StatusCode, e.Type, e.Detail)
}

// TOSError represents the error which is returned if the user needs to
// accept the TOS.
// TODO: include the new TOS url if we can somehow obtain it.
type TOSError struct {
	RemoteError
}

func handleHTTPError(resp *http.Response) error {
	var errorDetail RemoteError
	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&errorDetail)
	if err != nil {
		return err
	}

	errorDetail.StatusCode = resp.StatusCode

	// Check for errors we handle specifically
	if errorDetail.StatusCode == http.StatusForbidden && errorDetail.Detail == tosAgreementError {
		return TOSError{errorDetail}
	}

	return errorDetail
}

type domainError struct {
	Domain string
	Error  error
}
