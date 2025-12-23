package internal

type APIError struct{}

func (a *APIError) Error() string {
	// TODO implement me
	panic("implement me")
}

type Record struct {
	ID         int64  `json:"id,omitempty"`
	DomainID   int64  `json:"domain,omitempty"`
	Name       string `json:"name,omitempty"`
	Type       string `json:"type,omitempty"`
	Value      string `json:"value,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	Annotation string `json:"annotation,omitempty"`
}

type Domain struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	IsInternal bool   `json:"is_internal,omitempty"`
	Annotation string `json:"annotation,omitempty"`
}
