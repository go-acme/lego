package acme

import "time"

type registrationMessage struct {
	Resource string   `json:"resource"`
	Contact  []string `json:"contact"`
}

// Registration is returned by the ACME server after the registration
// The client implementation should save this registration somewhere.
type Registration struct {
	Resource string `json:"resource,omitempty"`
	ID       int    `json:"id"`
	Key      struct {
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"key"`
	Contact        []string `json:"contact"`
	Agreement      string   `json:"agreement,omitempty"`
	Authorizations string   `json:"authorizations,omitempty"`
	Certificates   string   `json:"certificates,omitempty"`
}

// RegistrationResource represents all important informations about a registration
// of which the client needs to keep track itself.
type RegistrationResource struct {
	Body        Registration
	URI         string
	NewAuthzURL string
	TosURL      string
}

type authorizationResource struct {
	Body       authorization
	Domain     string
	NewCertURL string
	AuthURL    string
}

type authorization struct {
	Resource     string      `json:"resource,omitempty"`
	Identifier   identifier  `json:"identifier"`
	Status       string      `json:"status,omitempty"`
	Expires      time.Time   `json:"expires,omitempty"`
	Challenges   []challenge `json:"challenges,omitempty"`
	Combinations [][]int     `json:"combinations,omitempty"`
}

type identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type challenge struct {
	Resource string `json:"resource,omitempty"`
	Type     string `json:"type,omitempty"`
	Status   string `json:"status,omitempty"`
	URI      string `json:"uri,omitempty"`
	Token    string `json:"token,omitempty"`
	Tls      bool   `json:"tls,omitempty"`
}

type csrMessage struct {
	Resource       string   `json:"resource,omitempty"`
	Csr            string   `json:"csr"`
	Authorizations []string `json:"authorizations"`
}

// CertificateResource represents a CA issued certificate.
// PrivateKey and Certificate are both already PEM encoded
// and can be directly written to disk.
type CertificateResource struct {
	Domain      string
	CertURL     string
	PrivateKey  []byte
	Certificate []byte
}
