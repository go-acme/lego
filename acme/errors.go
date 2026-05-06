package acme

import (
	"fmt"
	"strings"
	"time"
)

// ACME Error Types.
// - https://www.iana.org/assignments/acme/acme.xhtml#acme-error-types
const (
	errNS = "urn:ietf:params:acme:error:"

	// InvalidProfileErrorType The request specified a profile.
	// - https://datatracker.ietf.org/doc/html/draft-ietf-acme-profiles-01#name-acme-error-types
	InvalidProfileErrorType = errNS + "invalidProfile"

	// AlreadyReplacedErrorType The request specified a predecessor certificate that has already been marked as replaced
	// - https://www.rfc-editor.org/rfc/rfc9773.html#acme-error-types
	AlreadyReplacedErrorType = errNS + "alreadyReplaced"

	// AccountDoesNotExistErrorType The request specified an account that does not exist.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	AccountDoesNotExistErrorType = errNS + "accountDoesNotExist"

	// AlreadyRevokedErrorType The request specified a certificate to be revoked that has already been revoked.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	AlreadyRevokedErrorType = errNS + "alreadyRevoked"

	// BadCSRErrorType The CSR is unacceptable (e.g., due to a short key).
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	BadCSRErrorType = errNS + "badCSR"

	// BadNonceErrorType The client sent an unacceptable anti-replay nonce.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	BadNonceErrorType = errNS + "badNonce"

	// BadPublicKeyErrorType The JWS was signed by a public key the server does not support.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	BadPublicKeyErrorType = errNS + "badPublicKey"

	// BadRevocationReasonErrorType The revocation reason provided is not allowed by the server.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	BadRevocationReasonErrorType = errNS + "badRevocationReason"

	// BadSignatureAlgorithmErrorType The JWS was signed with an algorithm the server does not support.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	BadSignatureAlgorithmErrorType = errNS + "badSignatureAlgorithm"

	// CaaErrorType Certification Authority Authorization (CAA) records forbid the CA from issuing a certificate.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	CaaErrorType = errNS + "caa"

	// CompoundErrorType Specific error conditions are indicated in the "subproblems" array.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	CompoundErrorType = errNS + "compound"

	// ConnectionErrorType The server could not connect to validation target.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	ConnectionErrorType = errNS + "connection"

	// DNSErrorType There was a problem with a DNS query during identifier validation.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	DNSErrorType = errNS + "dns"

	// ExternalAccountRequiredErrorType The request must include a value for the "externalAccountBinding" field.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	ExternalAccountRequiredErrorType = errNS + "externalAccountRequired"

	// IncorrectResponseErrorType Response received didn't match the challenge's requirements.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	IncorrectResponseErrorType = errNS + "incorrectResponse"

	// InvalidContactErrorType A contact URL for an account was invalid.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	InvalidContactErrorType = errNS + "invalidContact"

	// MalformedErrorType The request message was malformed.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	MalformedErrorType = errNS + "malformed"

	// OrderNotReadyErrorType The request attempted to finalize an order that is not ready to be finalized.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	OrderNotReadyErrorType = errNS + "orderNotReady"

	// RateLimitedErrorType The request exceeds a rate limit.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	RateLimitedErrorType = errNS + "rateLimited"

	// RejectedIdentifierErrorType The server will not issue certificates for the identifier.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	RejectedIdentifierErrorType = errNS + "rejectedIdentifier"

	// ServerInternalErrorType The server experienced an internal error.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	ServerInternalErrorType = errNS + "serverInternal"

	// TLSErrorType The server received a TLS error during validation.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	TLSErrorType = errNS + "tls"

	// UnauthorizedErrorType The client lacks sufficient authorization.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	UnauthorizedErrorType = errNS + "unauthorized"

	// UnsupportedContactErrorType A contact URL for an account used an unsupported protocol scheme.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	UnsupportedContactErrorType = errNS + "unsupportedContact"

	// UnsupportedIdentifierErrorType An identifier is of an unsupported type.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	UnsupportedIdentifierErrorType = errNS + "unsupportedIdentifier"

	// UserActionRequiredErrorType Visit the "instance" URL and take actions specified there.
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-6.7
	UserActionRequiredErrorType = errNS + "userActionRequired"
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
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "acme: error: %d", p.HTTPStatus)

	if p.Method != "" || p.URL != "" {
		_, _ = fmt.Fprintf(msg, " :: %s :: %s", p.Method, p.URL)
	}

	_, _ = fmt.Fprintf(msg, " :: %s :: %s", p.Type, p.Detail)

	for _, sub := range p.SubProblems {
		_, _ = fmt.Fprintf(msg, ", problem: %q :: %s", sub.Type, sub.Detail)
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

	RetryAfter time.Duration
}

func (e *RateLimitedError) Unwrap() error {
	return e.ProblemDetails
}
