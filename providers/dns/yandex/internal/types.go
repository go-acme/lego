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

type AddResponse struct {
	Domain  string  `json:"domain,omitempty"`
	Record  *Record `json:"record,omitempty"`
	Success string  `json:"success"`
	Error   string  `json:"error,omitempty"`
}

type RemoveResponse struct {
	Domain   string `json:"domain,omitempty"`
	RecordID int    `json:"record_id,omitempty"`
	Success  string `json:"success"`
	Error    string `json:"error,omitempty"`
}

type ListResponse struct {
	Domain  string   `json:"domain,omitempty"`
	Records []Record `json:"records,omitempty"`
	Success string   `json:"success"`
	Error   string   `json:"error,omitempty"`
}
