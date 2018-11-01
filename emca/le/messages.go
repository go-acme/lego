// Package le contains all objects related the ACME endpoints.
// https://tools.ietf.org/html/draft-ietf-acme-acme-16
package le

import (
	"encoding/json"
	"time"
)

// Challenge statuses
// https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.1.6
const (
	StatusPending     = "pending"
	StatusInvalid     = "invalid"
	StatusValid       = "valid"
	StatusProcessing  = "processing"
	StatusDeactivated = "deactivated"
)

// Directory the ACME directory object.
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.1.1
type Directory struct {
	NewNonceURL   string `json:"newNonce"`
	NewAccountURL string `json:"newAccount"`
	NewOrderURL   string `json:"newOrder"`
	NewAuthzURL   string `json:"newAuthz"`
	RevokeCertURL string `json:"revokeCert"`
	KeyChangeURL  string `json:"keyChange"`
	Meta          Meta   `json:"meta"`
}

// Meta the ACME meta object (related to Directory).
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.1.1
type Meta struct {
	TermsOfService          string   `json:"termsOfService"`
	Website                 string   `json:"website"`
	CaaIdentities           []string `json:"caaIdentities"`
	ExternalAccountRequired bool     `json:"externalAccountRequired"`
}

// AccountMessage the ACME account Object.
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.1.2
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.3
type AccountMessage struct {
	Status                 string          `json:"status,omitempty"`
	Contact                []string        `json:"contact,omitempty"`
	TermsOfServiceAgreed   bool            `json:"termsOfServiceAgreed,omitempty"`
	Orders                 string          `json:"orders,omitempty"`
	OnlyReturnExisting     bool            `json:"onlyReturnExisting,omitempty"`
	ExternalAccountBinding json.RawMessage `json:"externalAccountBinding,omitempty"`
}

// OrderMessage the ACME order Object.
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.1.3
type OrderMessage struct {
	Status         string          `json:"status,omitempty"`
	Expires        string          `json:"expires,omitempty"`
	Identifiers    []Identifier    `json:"identifiers"`
	NotBefore      string          `json:"notBefore,omitempty"`
	NotAfter       string          `json:"notAfter,omitempty"`
	Error          *ProblemDetails `json:"error,omitempty"` // TODO new field
	Authorizations []string        `json:"authorizations,omitempty"`
	Finalize       string          `json:"finalize,omitempty"`
	Certificate    string          `json:"certificate,omitempty"`
}

// Authorization the ACME authorization object.
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.1.4
type Authorization struct {
	Status     string      `json:"status"`
	Expires    time.Time   `json:"expires,omitempty"`
	Identifier Identifier  `json:"identifier,omitempty"`
	Challenges []Challenge `json:"challenges,omitempty"`
	Wildcard   bool        `json:"wildcard,omitempty"`
}

// Challenge the ACME challenge object.
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.1.5
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-8
type Challenge struct {
	Type      string          `json:"type"`
	URL       string          `json:"url"`
	Status    string          `json:"status"`
	Validated time.Time       `json:"validated,omitempty"`
	Error     *ProblemDetails `json:"error,omitempty"`

	Token            string `json:"token"`
	KeyAuthorization string `json:"keyAuthorization"`
}

// Identifier the ACME identifier object.
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-9.7.7
type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CSRMessage Certificate Signing Request
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.4
type CSRMessage struct {
	Csr string `json:"csr"`
}

// RevokeCertMessage a certificate revocation message
// - https://tools.ietf.org/html/draft-ietf-acme-acme-16#section-7.6
// - https://tools.ietf.org/html/rfc5280#section-5.3.1
type RevokeCertMessage struct {
	Certificate string `json:"certificate"`
	Reason      *uint  `json:"reason,omitempty"` // TODO new field
}
