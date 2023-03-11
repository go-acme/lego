package internal

type Record struct {
	ID        int    `json:"record_id,omitempty" url:"record_id,omitempty"`
	Domain    string `json:"domain,omitempty" url:"domain,omitempty"`
	SubDomain string `json:"subdomain,omitempty" url:"subdomain,omitempty"`
	FQDN      string `json:"fqdn,omitempty" url:"fqdn,omitempty"`
	TTL       int    `json:"ttl,omitempty" url:"ttl,omitempty"`
	Type      string `json:"type,omitempty" url:"type,omitempty"`
	Content   string `json:"content,omitempty" url:"content,omitempty"`
}

type Response interface {
	GetSuccess() string
	GetError() string
}

type BaseResponse struct {
	Success string `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (r BaseResponse) GetSuccess() string {
	return r.Success
}

func (r BaseResponse) GetError() string {
	return r.Error
}

type AddResponse struct {
	BaseResponse
	Domain string  `json:"domain,omitempty"`
	Record *Record `json:"record,omitempty"`
}

type RemoveResponse struct {
	BaseResponse
	Domain   string `json:"domain,omitempty"`
	RecordID int    `json:"record_id,omitempty"`
}

type ListResponse struct {
	BaseResponse
	Domain  string   `json:"domain,omitempty"`
	Records []Record `json:"records,omitempty"`
}
