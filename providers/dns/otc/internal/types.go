package internal

// LoginRequest

type LoginRequest struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	Identity Identity `json:"identity"`
	Scope    Scope    `json:"scope"`
}

type Identity struct {
	Methods  []string `json:"methods"`
	Password Password `json:"password"`
}

type Password struct {
	User User `json:"user"`
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Domain   Domain `json:"domain"`
}

type Scope struct {
	Project Project `json:"project"`
}

type Project struct {
	Name string `json:"name"`
}

// TokenResponse

type TokenResponse struct {
	Token Token `json:"token"`
}

type Token struct {
	User      UserR     `json:"user,omitempty"`
	Domain    Domain    `json:"domain,omitempty"`
	Catalog   []Catalog `json:"catalog,omitempty"`
	Methods   []string  `json:"methods,omitempty"`
	Roles     []Role    `json:"roles,omitempty"`
	ExpiresAt string    `json:"expires_at,omitempty"`
	IssuedAt  string    `json:"issued_at,omitempty"`
}

type Catalog struct {
	ID        string     `json:"id,omitempty"`
	Type      string     `json:"type,omitempty"`
	Name      string     `json:"name,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

type UserR struct {
	ID                string `json:"id,omitempty"`
	Domain            Domain `json:"domain,omitempty"`
	Name              string `json:"name,omitempty"`
	PasswordExpiresAt string `json:"password_expires_at,omitempty"`
}

type Endpoint struct {
	ID        string `json:"id,omitempty"`
	URL       string `json:"url,omitempty"`
	Region    string `json:"region,omitempty"`
	RegionID  string `json:"region_id,omitempty"`
	Interface string `json:"interface,omitempty"`
}

type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// RecordSetsResponse

type RecordSetsResponse struct {
	Links      Links        `json:"links"`
	RecordSets []RecordSets `json:"recordsets"`
	Metadata   Metadata     `json:"metadata"`
}

type RecordSets struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	TTL         int      `json:"ttl,omitempty"`
	Records     []string `json:"records,omitempty"`

	Status    string `json:"status,omitempty"`
	Links     *Links `json:"links,omitempty"`
	ZoneID    string `json:"zone_id,omitempty"`
	ZoneName  string `json:"zone_name,omitempty"`
	CreateAt  string `json:"create_at,omitempty"`
	UpdateAt  string `json:"update_at,omitempty"`
	Default   bool   `json:"default,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// ZonesResponse

type ZonesResponse struct {
	Links    Links    `json:"links,omitempty"`
	Zones    []Zone   `json:"zones"`
	Metadata Metadata `json:"metadata"`
}

type Zone struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Email       string `json:"email,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	Serial      int    `json:"serial,omitempty"`
	Status      string `json:"status,omitempty"`
	Links       *Links `json:"links,omitempty"`
	PoolID      string `json:"pool_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ZoneType    string `json:"zone_type,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	RecordNum   int    `json:"record_num,omitempty"`
}

// Response

type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
}

type Metadata struct {
	TotalCount int `json:"total_count,omitempty"`
}

// Shared

type Domain struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
