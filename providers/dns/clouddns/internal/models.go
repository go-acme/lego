package internal

type APIError struct {
	Error ErrorContent `json:"error"`
}

type ErrorContent struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Authorization struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type AuthResponse struct {
	Auth Auth `json:"auth,omitempty"`
}

type Auth struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type SearchQuery struct {
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
	Search []Search `json:"search,omitempty"`
	Sort   []Sort   `json:"sort,omitempty"`
}

// Search used for searches in the CloudDNS API.
type Search struct {
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Type     string `json:"type"`
	Value    string `json:"value"`
}

type Sort struct {
	Ascending bool   `json:"ascending"`
	Name      string `json:"name"`
}

type SearchResponse struct {
	Items     []Domain `json:"items"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
	TotalHits int      `json:"totalHits"`
}

type Domain struct {
	ID         string `json:"id"`
	DomainName string `json:"domainName"`
	Status     string `json:"status"`
}

// Record represents a DNS record.
type Record struct {
	ID       string `json:"id,omitempty"`
	DomainID string `json:"domainId,omitempty"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Type     string `json:"type,omitempty"`
}

type DomainInfo struct {
	ID                   string   `json:"id,omitempty"`
	DomainName           string   `json:"domainName,omitempty"`
	LastDomainRecordList []Record `json:"lastDomainRecordList,omitempty"`
	SoaTTL               int      `json:"soaTtl,omitempty"`
	Status               string   `json:"status,omitempty"`
}
