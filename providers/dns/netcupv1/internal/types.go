package internal

import "fmt"

const (
	OrderByDomainID = "domainId"
	OrderByFQDN     = "fqdn"
)

const (
	OrderASC  = "asc"
	OrderDESC = "desc"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}

type APIResponse[T any] struct {
	Success  bool         `json:"success"`
	Errors   []*APIError  `json:"errors"`
	Messages []APIMessage `json:"messages"`
	Meta     *Metadata    `json:"meta"`
	Result   []T          `json:"result"`
}

type APIMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Metadata struct {
	Sort       *PageSort   `json:"sort"`
	Pagination *Pagination `json:"pagination"`
}

type PageSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type Pagination struct {
	Page         int `json:"page"`
	PerPage      int `json:"perPage"`
	TotalEntries int `json:"totalEntries"`
	PreviousPage int `json:"previousPage"`
	NextPage     int `json:"nextPage"`
	LastPage     int `json:"lastPage"`
}

type Domain struct {
	ID           int64  `json:"id"`
	FQDN         string `json:"fqdn"`
	IsDNSManaged bool   `json:"isDnsManaged"`
}

type ChallengeData struct {
	Scope  string `json:"scope,omitempty"`
	FQDN   string `json:"fqdn,omitempty"`
	Value  string `json:"value,omitempty"`
	Status string `json:"status,omitempty"`
}

type DomainsRequest struct {
	Field     string `url:"field,omitempty"`
	Direction string `url:"direction,omitempty"`
	Page      int    `url:"page,omitempty"`
	PerPage   int    `url:"perPage,omitempty"`
	FQDN      string `url:"fqdn,omitempty"`
	DomainID  string `url:"domainId,omitempty"`
}
