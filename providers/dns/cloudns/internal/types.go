package internal

type apiResponse struct {
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
}

// Zone is a zone.
type Zone struct {
	Name   string
	Type   string
	Zone   string
	Status string // is an integer, but cast as string
}

// TXTRecord is a TXT record.
type TXTRecord struct {
	ID       int    `json:"id,string"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Record   string `json:"record"`
	Failover int    `json:"failover,string"`
	TTL      int    `json:"ttl,string"`
	Status   int    `json:"status"`
}

// UpdateRecord is a Server Sync Record.
type UpdateRecord struct {
	Server  string `json:"server"`
	IP4     string `json:"ip4"`
	IP6     string `json:"ip6"`
	Updated bool   `json:"updated"`
}

type SyncProgress struct {
	Complete bool
	Updated  int
	Total    int
}
