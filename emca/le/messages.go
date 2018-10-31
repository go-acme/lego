package le

import (
	"encoding/json"
	"time"
)

// RegistrationResource represents all important information about a registration
// of which the client needs to keep track itself.
type RegistrationResource struct {
	Body AccountMessage `json:"body,omitempty"`
	URI  string         `json:"uri,omitempty"`
}

// CertificateResource represents a CA issued certificate.
// PrivateKey, Certificate and IssuerCertificate are all
// already PEM encoded and can be directly written to disk.
// Certificate may be a certificate bundle, depending on the
// options supplied to create it.
type CertificateResource struct {
	Domain            string `json:"domain"`
	CertURL           string `json:"certUrl"`
	CertStableURL     string `json:"certStableUrl"`
	AccountRef        string `json:"accountRef,omitempty"`
	PrivateKey        []byte `json:"-"`
	Certificate       []byte `json:"-"`
	IssuerCertificate []byte `json:"-"`
	CSR               []byte `json:"-"`
}

type Challenge struct {
	URL              string       `json:"url"`
	Type             string       `json:"type"`
	Status           string       `json:"status"`
	Token            string       `json:"token"`
	Validated        time.Time    `json:"validated"`
	KeyAuthorization string       `json:"keyAuthorization"`
	Error            ErrorDetails `json:"error"`
}

type Directory struct {
	NewNonceURL   string `json:"newNonce"`
	NewAccountURL string `json:"newAccount"`
	NewOrderURL   string `json:"newOrder"`
	RevokeCertURL string `json:"revokeCert"`
	KeyChangeURL  string `json:"keyChange"`
	Meta          struct {
		TermsOfService          string   `json:"termsOfService"`
		Website                 string   `json:"website"`
		CaaIdentities           []string `json:"caaIdentities"`
		ExternalAccountRequired bool     `json:"externalAccountRequired"`
	} `json:"meta"`
}

type AccountMessage struct {
	Status                 string          `json:"status,omitempty"`
	Contact                []string        `json:"contact,omitempty"`
	TermsOfServiceAgreed   bool            `json:"termsOfServiceAgreed,omitempty"`
	Orders                 string          `json:"orders,omitempty"`
	OnlyReturnExisting     bool            `json:"onlyReturnExisting,omitempty"`
	ExternalAccountBinding json.RawMessage `json:"externalAccountBinding,omitempty"`
}

type OrderResource struct {
	URL          string   `json:"url,omitempty"`
	Domains      []string `json:"domains,omitempty"`
	OrderMessage `json:"body,omitempty"`
}

type OrderMessage struct {
	Status         string       `json:"status,omitempty"`
	Expires        string       `json:"expires,omitempty"`
	Identifiers    []Identifier `json:"identifiers"`
	NotBefore      string       `json:"notBefore,omitempty"`
	NotAfter       string       `json:"notAfter,omitempty"`
	Authorizations []string     `json:"authorizations,omitempty"`
	Finalize       string       `json:"finalize,omitempty"`
	Certificate    string       `json:"certificate,omitempty"`
}

type Authorization struct {
	Status     string      `json:"status"`
	Expires    time.Time   `json:"expires"`
	Identifier Identifier  `json:"identifier"`
	Challenges []Challenge `json:"challenges"`
}

type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type CSRMessage struct {
	Csr string `json:"csr"`
}

type RevokeCertMessage struct {
	Certificate string `json:"certificate"`
}

type DeactivateAuthMessage struct {
	Status string `json:"status"`
}
