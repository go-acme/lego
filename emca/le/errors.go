package le

import (
	"fmt"
)

const (
	errNS       = "urn:ietf:params:acme:error:"
	BadNonceErr = errNS + "badNonce"
)

type ErrorDetails struct {
	Type       string `json:"type,omitempty"`
	Detail     string `json:"detail,omitempty"`
	HTTPStatus int    `json:"status,omitempty"`
	Method     string `json:"method,omitempty"`
	URL        string `json:"url,omitempty"`
}

func (p ErrorDetails) Error() string {
	return fmt.Sprintf("LE error: %d :: %s :: %s :: %s :: %s", p.HTTPStatus, p.Method, p.URL, p.Type, p.Detail)
}

// NonceError represents the error which is returned
// if the nonce sent by the client was not accepted by the server.
type NonceError struct {
	*ErrorDetails
}
