package internal

// IdentityRequest is an authentication request body.
type IdentityRequest struct {
	Auth Auth `json:"auth"`
}

// Auth is an authentication information.
type Auth struct {
	TenantID            string              `json:"tenantId"`
	PasswordCredentials PasswordCredentials `json:"passwordCredentials"`
}

// PasswordCredentials is API-user's credentials.
type PasswordCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// IdentityResponse is an authentication response body.
type IdentityResponse struct {
	Access Access `json:"access"`
}

// Access is an identity information.
type Access struct {
	Token Token `json:"token"`
}

// Token is an api access token.
type Token struct {
	ID string `json:"id"`
}

// DomainListResponse is a response of a domain listing request.
type DomainListResponse struct {
	Domains []Domain `json:"domains"`
}

// Domain is a hosted domain entry.
type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecordListResponse is a response of record listing request.
type RecordListResponse struct {
	Records []Record `json:"records"`
}

// Record is a record entry.
type Record struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}
