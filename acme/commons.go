// Package acme contains all objects related the ACME endpoints.
// https://www.rfc-editor.org/rfc/rfc8555.html
package acme

import (
	"encoding/json"
	"time"
)

// ACME status values of Account, Order, Authorization and Challenge objects.
// See https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.6 for details.
const (
	StatusDeactivated = "deactivated"
	StatusExpired     = "expired"
	StatusInvalid     = "invalid"
	StatusPending     = "pending"
	StatusProcessing  = "processing"
	StatusReady       = "ready"
	StatusRevoked     = "revoked"
	StatusUnknown     = "unknown"
	StatusValid       = "valid"
)

// CRL reason codes as defined in RFC 5280.
// https://datatracker.ietf.org/doc/html/rfc5280#section-5.3.1
const (
	CRLReasonUnspecified          uint = 0
	CRLReasonKeyCompromise        uint = 1
	CRLReasonCACompromise         uint = 2
	CRLReasonAffiliationChanged   uint = 3
	CRLReasonSuperseded           uint = 4
	CRLReasonCessationOfOperation uint = 5
	CRLReasonCertificateHold      uint = 6
	CRLReasonRemoveFromCRL        uint = 8
	CRLReasonPrivilegeWithdrawn   uint = 9
	CRLReasonAACompromise         uint = 10
)

// Directory the ACME directory object.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.1
// - https://www.rfc-editor.org/rfc/rfc9773.html
type Directory struct {
	NewNonceURL   string `json:"newNonce"`
	NewAccountURL string `json:"newAccount"`
	NewOrderURL   string `json:"newOrder"`
	NewAuthzURL   string `json:"newAuthz"`
	RevokeCertURL string `json:"revokeCert"`
	KeyChangeURL  string `json:"keyChange"`
	Meta          Meta   `json:"meta"`
	RenewalInfo   string `json:"renewalInfo"`
}

// Meta the ACME meta object (related to Directory).
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.1
type Meta struct {
	// termsOfService (optional, string):
	// A URL identifying the current terms of service.
	TermsOfService string `json:"termsOfService"`

	// website (optional, string):
	// An HTTP or HTTPS URL locating a website providing more information about the ACME server.
	Website string `json:"website"`

	// caaIdentities (optional, array of string):
	// The hostnames that the ACME server recognizes as referring to itself
	// for the purposes of CAA record validation as defined in [RFC6844].
	// Each string MUST represent the same sequence of ASCII code points
	// that the server will expect to see as the "Issuer Domain Name" in a CAA issue or issuewild property tag.
	// This allows clients to determine the correct issuer domain name to use when configuring CAA records.
	CaaIdentities []string `json:"caaIdentities"`

	// externalAccountRequired (optional, boolean):
	// If this field is present and set to "true",
	// then the CA requires that all new-account requests include an "externalAccountBinding" field
	// associating the new account with an external account.
	ExternalAccountRequired bool `json:"externalAccountRequired"`

	// profiles (optional, object):
	// A map of profile names to human-readable descriptions of those profiles.
	// https://www.ietf.org/id/draft-aaron-acme-profiles-00.html#section-3
	Profiles map[string]string `json:"profiles"`
}

// ExtendedAccount an extended Account.
type ExtendedAccount struct {
	Account
	// Contains the value of the response header `Location`
	Location string `json:"-"`
}

// Account the ACME account Object.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.2
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.3
type Account struct {
	// status (required, string):
	// The status of this account.
	// Possible values are: "valid", "deactivated", and "revoked".
	// The value "deactivated" should be used to indicate client-initiated deactivation
	// whereas "revoked" should be used to indicate server-initiated deactivation. (See Section 7.1.6)
	Status string `json:"status,omitempty"`

	// contact (optional, array of string):
	// An array of URLs that the server can use to contact the client for issues related to this account.
	// For example, the server may wish to notify the client about server-initiated revocation or certificate expiration.
	// For information on supported URL schemes, see Section 7.3
	Contact []string `json:"contact,omitempty"`

	// termsOfServiceAgreed (optional, boolean):
	// Including this field in a new-account request,
	// with a value of true, indicates the client's agreement with the terms of service.
	// This field is not updateable by the client.
	TermsOfServiceAgreed bool `json:"termsOfServiceAgreed,omitempty"`

	// orders (required, string):
	// A URL from which a list of orders submitted by this account can be fetched via a POST-as-GET request,
	// as described in Section 7.1.2.1.
	Orders string `json:"orders,omitempty"`

	// onlyReturnExisting (optional, boolean):
	// If this field is present with the value "true",
	// then the server MUST NOT create a new account if one does not already exist.
	// This allows a client to look up an account URL based on an account key (see Section 7.3.1).
	OnlyReturnExisting bool `json:"onlyReturnExisting,omitempty"`

	// externalAccountBinding (optional, object):
	// An optional field for binding the new account with an existing non-ACME account (see Section 7.3.4).
	ExternalAccountBinding json.RawMessage `json:"externalAccountBinding,omitempty"`
}

// ExtendedOrder a extended Order.
type ExtendedOrder struct {
	Order

	// The order URL, contains the value of the response header `Location`
	Location string `json:"-"`
}

// Order the ACME order Object.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.3
type Order struct {
	// status (required, string):
	// The status of this order.
	// Possible values are: "pending", "ready", "processing", "valid", and "invalid".
	Status string `json:"status,omitempty"`

	// expires (optional, string):
	// The timestamp after which the server will consider this order invalid,
	// encoded in the format specified in RFC 3339 [RFC3339].
	// This field is REQUIRED for objects with "pending" or "valid" in the status field.
	Expires string `json:"expires,omitempty"`

	// identifiers (required, array of object):
	// An array of identifier objects that the order pertains to.
	Identifiers []Identifier `json:"identifiers"`

	// profile (string, optional):
	// A string uniquely identifying the profile
	// which will be used to affect issuance of the certificate requested by this Order.
	// https://www.ietf.org/id/draft-aaron-acme-profiles-00.html#section-4
	Profile string `json:"profile,omitempty"`

	// notBefore (optional, string):
	// The requested value of the notBefore field in the certificate,
	// in the date format defined in [RFC3339].
	NotBefore string `json:"notBefore,omitempty"`

	// notAfter (optional, string):
	// The requested value of the notAfter field in the certificate,
	// in the date format defined in [RFC3339].
	NotAfter string `json:"notAfter,omitempty"`

	// error (optional, object):
	// The error that occurred while processing the order, if any.
	// This field is structured as a problem document [RFC7807].
	Error *ProblemDetails `json:"error,omitempty"`

	// authorizations (required, array of string):
	// For pending orders,
	// the authorizations that the client needs to complete before the requested certificate can be issued (see Section 7.5),
	// including unexpired authorizations that the client has completed in the past for identifiers specified in the order.
	// The authorizations required are dictated by server policy
	// and there may not be a 1:1 relationship between the order identifiers and the authorizations required.
	// For final orders (in the "valid" or "invalid" state), the authorizations that were completed.
	// Each entry is a URL from which an authorization can be fetched with a POST-as-GET request.
	Authorizations []string `json:"authorizations,omitempty"`

	// finalize (required, string):
	// A URL that a CSR must be POSTed to once all of the order's authorizations are satisfied to finalize the order.
	// The result of a successful finalization will be the population of the certificate URL for the order.
	Finalize string `json:"finalize,omitempty"`

	// certificate (optional, string):
	// A URL for the certificate that has been issued in response to this order
	Certificate string `json:"certificate,omitempty"`

	// replaces (optional, string):
	// replaces (string, optional): A string uniquely identifying a
	// previously-issued certificate which this order is intended to replace.
	// - https://www.rfc-editor.org/rfc/rfc9773.html#section-5
	Replaces string `json:"replaces,omitempty"`
}

func (r *Order) Err() error {
	if r.Error != nil {
		return r.Error
	}

	return nil
}

// Authorization the ACME authorization object.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.4
type Authorization struct {
	// status (required, string):
	// The status of this authorization.
	// Possible values are: "pending", "valid", "invalid", "deactivated", "expired", and "revoked".
	Status string `json:"status"`

	// expires (optional, string):
	// The timestamp after which the server will consider this authorization invalid,
	// encoded in the format specified in RFC 3339 [RFC3339].
	// This field is REQUIRED for objects with "valid" in the "status" field.
	Expires time.Time `json:"expires,omitempty"`

	// identifier (required, object):
	// The identifier that the account is authorized to represent
	Identifier Identifier `json:"identifier,omitempty"`

	// challenges (required, array of objects):
	// For pending authorizations, the challenges that the client can fulfill in order to prove possession of the identifier.
	// For valid authorizations, the challenge that was validated.
	// For invalid authorizations, the challenge that was attempted and failed.
	// Each array entry is an object with parameters required to validate the challenge.
	// A client should attempt to fulfill one of these challenges,
	// and a server should consider any one of the challenges sufficient to make the authorization valid.
	Challenges []Challenge `json:"challenges,omitempty"`

	// wildcard (optional, boolean):
	// For authorizations created as a result of a newOrder request containing a DNS identifier
	// with a value that contained a wildcard prefix this field MUST be present, and true.
	Wildcard bool `json:"wildcard,omitempty"`
}

// ExtendedChallenge a extended Challenge.
type ExtendedChallenge struct {
	Challenge
	// Contains the value of the response header `Retry-After`
	RetryAfter string `json:"-"`
	// Contains the value of the response header `Link` rel="up"
	AuthorizationURL string `json:"-"`
}

// Challenge the ACME challenge object.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.1.5
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-8
type Challenge struct {
	// type (required, string):
	// The type of challenge encoded in the object.
	Type string `json:"type"`

	// url (required, string):
	// The URL to which a response can be posted.
	URL string `json:"url"`

	// status (required, string):
	// The status of this challenge. Possible values are: "pending", "processing", "valid", and "invalid".
	Status string `json:"status"`

	// validated (optional, string):
	// The time at which the server validated this challenge,
	// encoded in the format specified in RFC 3339 [RFC3339].
	// This field is REQUIRED if the "status" field is "valid".
	Validated time.Time `json:"validated,omitempty"`

	// error (optional, object):
	// Error that occurred while the server was validating the challenge, if any,
	// structured as a problem document [RFC7807].
	// Multiple errors can be indicated by using subproblems Section 6.7.1.
	// A challenge object with an error MUST have status equal to "invalid".
	Error *ProblemDetails `json:"error,omitempty"`

	// token (required, string):
	// A random value that uniquely identifies the challenge.
	// This value MUST have at least 128 bits of entropy.
	// It MUST NOT contain any characters outside the base64url alphabet,
	// and MUST NOT include base64 padding characters ("=").
	// See [RFC4086] for additional information on randomness requirements.
	// https://www.rfc-editor.org/rfc/rfc8555.html#section-8.3
	// https://www.rfc-editor.org/rfc/rfc8555.html#section-8.4
	Token string `json:"token"`

	// https://www.rfc-editor.org/rfc/rfc8555.html#section-8.1
	KeyAuthorization string `json:"keyAuthorization"`
}

func (c *Challenge) Err() error {
	if c.Error != nil {
		return c.Error
	}

	return nil
}

// Identifier the ACME identifier object.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-9.7.7
type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CSRMessage Certificate Signing Request.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.4
type CSRMessage struct {
	// csr (required, string):
	// A CSR encoding the parameters for the certificate being requested [RFC2986].
	// The CSR is sent in the base64url-encoded version of the DER format.
	// (Note: Because this field uses base64url, and does not include headers, it is different from PEM.).
	Csr string `json:"csr"`
}

// RevokeCertMessage a certificate revocation message.
// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.6
// - https://www.rfc-editor.org/rfc/rfc5280.html#section-5.3.1
type RevokeCertMessage struct {
	// certificate (required, string):
	// The certificate to be revoked, in the base64url-encoded version of the DER format.
	// (Note: Because this field uses base64url, and does not include headers, it is different from PEM.)
	Certificate string `json:"certificate"`

	// reason (optional, int):
	// One of the revocation reasonCodes defined in Section 5.3.1 of [RFC5280] to be used when generating OCSP responses and CRLs.
	// If this field is not set the server SHOULD omit the reasonCode CRL entry extension when generating OCSP responses and CRLs.
	// The server MAY disallow a subset of reasonCodes from being used by the user.
	// If a request contains a disallowed reasonCode the server MUST reject it with the error type "urn:ietf:params:acme:error:badRevocationReason".
	// The problem document detail SHOULD indicate which reasonCodes are allowed.
	Reason *uint `json:"reason,omitempty"`
}

// RawCertificate raw data of a certificate.
type RawCertificate struct {
	Cert   []byte
	Issuer []byte
}

// Window is a window of time.
type Window struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// RenewalInfoResponse is the response to GET requests made the renewalInfo endpoint.
// - (4.1. Getting Renewal Information) https://www.rfc-editor.org/rfc/rfc9773.html
type RenewalInfoResponse struct {
	// SuggestedWindow contains two fields, start and end,
	// whose values are timestamps which bound the window of time in which the CA recommends renewing the certificate.
	SuggestedWindow Window `json:"suggestedWindow"`
	//	ExplanationURL is an optional URL pointing to a page which may explain why the suggested renewal window is what it is.
	//	For example, it may be a page explaining the CA's dynamic load-balancing strategy,
	//	or a page documenting which certificates are affected by a mass revocation event.
	//	Callers SHOULD provide this URL to their operator, if present.
	ExplanationURL string `json:"explanationURL"`
}

// RenewalInfoUpdateRequest is the JWS payload for POST requests made to the renewalInfo endpoint.
// - (4.2. RenewalInfo Objects) https://www.rfc-editor.org/rfc/rfc9773.html#section-4.2
type RenewalInfoUpdateRequest struct {
	// CertID is a composite string in the format: base64url(AKI) || '.' || base64url(Serial), where AKI is the
	// certificate's authority key identifier and Serial is the certificate's serial number. For details, see:
	// https://www.rfc-editor.org/rfc/rfc9773.html#section-4.1
	CertID string `json:"certID"`
	// Replaced is required and indicates whether or not the client considers the certificate to have been replaced.
	// A certificate is considered replaced when its revocation would not disrupt any ongoing services,
	// for instance because it has been renewed and the new certificate is in use, or because it is no longer in use.
	// Clients SHOULD NOT send a request where this value is false.
	Replaced bool `json:"replaced"`
}
