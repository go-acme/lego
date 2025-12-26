package internal

type RecordRequest struct {
	ID            int64  `json:"id,omitempty"`
	DomainID      int64  `json:"domain,omitempty"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	Value         string `json:"value,omitempty"`
	TTL           int    `json:"ttl,omitempty"`
	Annotation    string `json:"annotation,omitempty"`
	IsUserDefined bool   `json:"is_user_defined,omitempty"`
	IsActive      bool   `json:"is_active,omitempty"`
}

type Record struct {
	ID            int64   `json:"id,omitempty"`
	Domain        *Domain `json:"domain,omitempty"`
	Type          string  `json:"type,omitempty"`
	Name          string  `json:"name,omitempty"`
	Value         string  `json:"value,omitempty"`
	TTL           int     `json:"ttl,omitempty"`
	Annotation    string  `json:"annotation,omitempty"`
	IsUserDefined bool    `json:"is_user_defined,omitempty"`
	IsActive      bool    `json:"is_active,omitempty"`
}

type Domain struct {
	ID         int64  `json:"id,omitempty"`
	Href       string `json:"href,omitempty"`
	Name       string `json:"name,omitempty"`
	IsInternal bool   `json:"is_internal,omitempty"`
	Annotation string `json:"annotation,omitempty"`
}
