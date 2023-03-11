package internal

// Authentication response.

// Identity api structure.
type Identity struct {
	Access Access `json:"access"`
}

// Access api structure.
type Access struct {
	Token          Token            `json:"token"`
	ServiceCatalog []ServiceCatalog `json:"serviceCatalog"`
	User           User             `json:"user"`
}

// Token api structure.
type Token struct {
	ID                     string   `json:"id"`
	Expires                string   `json:"expires"`
	Tenant                 Tenant   `json:"tenant"`
	RAXAUTHAuthenticatedBy []string `json:"RAX-AUTH:authenticatedBy"`
}

// ServiceCatalog service catalog.
type ServiceCatalog struct {
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Tenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Endpoint api structure.
type Endpoint struct {
	PublicURL   string `json:"publicURL"`
	Region      string `json:"region,omitempty"`
	TenantID    string `json:"tenantId"`
	InternalURL string `json:"internalURL,omitempty"`
}

type Role struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	TenantID    string `json:"tenantId,omitempty"`
}

type User struct {
	ID                   string `json:"id"`
	Roles                []Role `json:"roles"`
	Name                 string `json:"name"`
	RAXAUTHDefaultRegion string `json:"RAX-AUTH:defaultRegion"`
}

// Authentication request.

// AuthData api structure.
type AuthData struct {
	Auth `json:"auth"`
}

// Auth api structure.
type Auth struct {
	APIKeyCredentials `json:"RAX-KSKEY:apiKeyCredentials"`
}

// APIKeyCredentials api structure.
type APIKeyCredentials struct {
	Username string `json:"username"`
	APIKey   string `json:"apiKey"`
}

// API responses.

// ZoneSearchResponse represents the response when querying Rackspace DNS zones.
type ZoneSearchResponse struct {
	TotalEntries int          `json:"totalEntries"`
	HostedZones  []HostedZone `json:"domains"`
}

// HostedZone api structure.
type HostedZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Records is the list of records sent/received from the DNS API.
type Records struct {
	TotalEntries int      `json:"totalEntries,omitempty"`
	Records      []Record `json:"records,omitempty"`
}

// Record represents a Rackspace DNS record.
type Record struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl,omitempty"`
	ID   string `json:"id,omitempty"`
}
