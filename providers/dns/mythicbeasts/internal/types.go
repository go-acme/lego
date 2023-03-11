package internal

import (
	"fmt"
	"time"
)

type Token struct {
	// The bearer token for use in API requests
	Token string `json:"access_token"`

	// The maximum lifetime of the token in seconds
	Lifetime int `json:"expires_in"`

	// The token type (must be 'bearer')
	TokenType string `json:"token_type"`

	Deadline time.Time `json:"-"`
}

type authResponseError struct {
	ErrorMsg         string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (a authResponseError) Error() string {
	return fmt.Sprintf("%s: %s", a.ErrorMsg, a.ErrorDescription)
}

type createTXTRequest struct {
	Records []createTXTRecord `json:"records"`
}

type createTXTRecord struct {
	Host string `json:"host"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
	Data string `json:"data"`
}

type createTXTResponse struct {
	Added   int    `json:"records_added"`
	Removed int    `json:"records_removed"`
	Message string `json:"message"`
}

type deleteTXTResponse struct {
	Removed int    `json:"records_removed"`
	Message string `json:"message"`
}
