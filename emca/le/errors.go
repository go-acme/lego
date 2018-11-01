package le

import (
	"fmt"
)

// Errors types
const (
	errNS       = "urn:ietf:params:acme:error:"
	BadNonceErr = errNS + "badNonce"
)

// ProblemDetails the problem details object
// - https://tools.ietf.org/html/rfc7807#section-3.1
type ProblemDetails struct {
	Type       string `json:"type,omitempty"`
	Detail     string `json:"detail,omitempty"`
	HTTPStatus int    `json:"status,omitempty"`

	// additional values to have a better error message (Not defined by the RFC)
	Method string `json:"method,omitempty"`
	URL    string `json:"url,omitempty"`
}

func (p ProblemDetails) Error() string {
	return fmt.Sprintf("LE error: %d :: %s :: %s :: %s :: %s", p.HTTPStatus, p.Method, p.URL, p.Type, p.Detail)
}

// NonceError represents the error which is returned
// if the nonce sent by the client was not accepted by the server.
type NonceError struct {
	*ProblemDetails
}
