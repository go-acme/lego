package acme

import (
	"fmt"
	"strings"
)

// Errors types.
const (
	errNS              = "urn:ietf:params:acme:error:"
	BadNonceErr        = errNS + "badNonce"
	AlreadyReplacedErr = errNS + "alreadyReplaced"
	RateLimitedErr     = errNS + "rateLimited"
)

// ProblemDetails the problem details object.
// - https://www.rfc-editor.org/rfc/rfc7807.html#section-3.1
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.3.3
type ProblemDetails struct {
	Type        string       `json:"type,omitempty"`
	Detail      string       `json:"detail,omitempty"`
	HTTPStatus  int          `json:"status,omitempty"`
	Instance    string       `json:"instance,omitempty"`
	SubProblems []SubProblem `json:"subproblems,omitempty"`

	// additional values to have a better error message (Not defined by the RFC)
	Method string `json:"method,omitempty"`
	URL    string `json:"url,omitempty"`
}

func (p *ProblemDetails) Error() string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("acme: error: %d", p.HTTPStatus))

	if p.Method != "" || p.URL != "" {
		msg.WriteString(fmt.Sprintf(" :: %s :: %s", p.Method, p.URL))
	}

	msg.WriteString(fmt.Sprintf(" :: %s :: %s", p.Type, p.Detail))

	for _, sub := range p.SubProblems {
		msg.WriteString(fmt.Sprintf(", problem: %q :: %s", sub.Type, sub.Detail))
	}

	if p.Instance != "" {
		msg.WriteString(", url: " + p.Instance)
	}

	return msg.String()
}

// SubProblem a "subproblems".
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7.1
type SubProblem struct {
	Type       string     `json:"type,omitempty"`
	Detail     string     `json:"detail,omitempty"`
	Identifier Identifier `json:"identifier"`
}

// NonceError represents the error which is returned
// if the nonce sent by the client was not accepted by the server.
type NonceError struct {
	*ProblemDetails
}

func (e *NonceError) Unwrap() error {
	return e.ProblemDetails
}

// AlreadyReplacedError represents the error which is returned
// if the Server rejects the request because the identified certificate has already been marked as replaced.
// - https://www.rfc-editor.org/rfc/rfc9773.html#section-5
type AlreadyReplacedError struct {
	*ProblemDetails
}

func (e *AlreadyReplacedError) Unwrap() error {
	return e.ProblemDetails
}

// RateLimitedError represents the error which is returned
// if the server rejects the request because the client has exceeded the rate limit.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.6
type RateLimitedError struct {
	*ProblemDetails

	RetryAfter string
}

func (e *RateLimitedError) Unwrap() error {
	return e.ProblemDetails
}
