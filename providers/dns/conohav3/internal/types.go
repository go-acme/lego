package internal

// IdentityRequest is the top-level payload sent to the Identity v3.
type IdentityRequest struct {
	Auth Auth `json:"auth"`
}

// Auth authentication credentials (Identity) and scope (Scope).
type Auth struct {
	Identity Identity `json:"identity"`
	Scope    Scope    `json:"scope"`
}

// Identity describes how the client will authenticate.
// In ConoHa v3.0, only support the "password" method.
type Identity struct {
	Methods  []string `json:"methods"`
	Password Password `json:"password"`
}

// Password nests the concrete user credentials used by the password auth method.
type Password struct {
	User User `json:"user"`
}

// User holds the API User ID and password that will be verified by the Identity service.
type User struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

// Scope specifies which tenant the issued token should be scoped to.
type Scope struct {
	Project Project `json:"project"`
}

// Project identifies the target tenant by UUID.
type Project struct {
	ID string `json:"id"`
}

// DomainListResponse is returned by `GET /v1/domains` and contains all DNS zones (domains) owned by the project.
type DomainListResponse struct {
	Domains []Domain `json:"domains"`
}

// Domain represents a single hosted DNS zone.
type Domain struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// RecordListResponse is returned by `GET /v1/domains/{domain_uuid}/records` and lists every record in the zone.
type RecordListResponse struct {
	Records []Record `json:"records"`
}

// Record represents a DNS record inside a zone.
type Record struct {
	UUID string `json:"uuid,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}
