package internal

const (
	SearchOperationBeginsWith = "bw"
	SearchOperationContains   = "cn"
	SearchOperationEndsWith   = "ew"
	SearchOperationEqual      = "eq"
	SearchOperationNotEqual   = "ne"
)

const (
	FieldDomainName = "domainName"
	FieldDomainType = "domainType"
)

type APIError struct {
	StatusCode string `json:"StatusCode"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

func (a *APIError) Error() string {
	// TODO implement me
	panic("implement me")
}

type APIResponse struct {
	ID      string `json:"id"`
	Options string `json:"options"`
	Data    []Data `json:"data"`

	Type       string `json:"Type"`
	StatusCode string `json:"StatusCode"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type Data struct {
	Type       string `json:"Type"`
	StatusCode string `json:"StatusCode"`
}

type TXTRecord struct {
	HostName string `json:"txtHostName"`
	Text     string `json:"txtText"`
	TTL      string `json:"txtTTL"`
}

type Options struct {
	PerPage         int    `url:"perpage,omitempty"`
	PageNumber      int    `url:"pageno,omitempty"`
	SortField       string `url:"sortField,omitempty"`
	SortOrder       string `url:"sortOrder,omitempty"`
	SearchField     string `url:"searchField,omitempty"`
	SearchString    string `url:"searchString,omitempty"`
	SearchOperation string `url:"searchOper,omitempty"`
}

type DomainsResponse struct {
	Rows []Domain `json:"rows"`

	Type       string `json:"Type"`
	StatusCode string `json:"StatusCode"`
}

type Domain struct {
	ID         string `json:"id"`
	DomainName string `json:"domainName"`
}
