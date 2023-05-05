package internal

import "fmt"

type APIError struct {
	Error ErrorContent `json:"error"`
}

type ErrorContent struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e ErrorContent) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type Authorization struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type AuthResponse struct {
	Auth Auth `json:"auth,omitempty"`
}

type Auth struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type SearchQuery struct {
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
	Search []Search `json:"search,omitempty"`
	Sort   []Sort   `json:"sort,omitempty"`
}

// Search used for searches in the CloudDNS API.
type Search struct {
	Name     string `json:"name,omitempty"`
	Operator string `json:"operator,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
}

type Sort struct {
	Ascending bool   `json:"ascending,omitempty"`
	Name      string `json:"name,omitempty"`
}

type SearchResponse struct {
	Items     []Domain `json:"items,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
	TotalHits int      `json:"totalHits,omitempty"`
}

type Domain struct {
	ID         string `json:"id,omitempty"`
	DomainName string `json:"domainName,omitempty"`
	Status     string `json:"status,omitempty"`
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
