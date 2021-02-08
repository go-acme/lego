package internal

// Domain JSON data structure.
type Domain struct {
	Name           string   `json:"domain"`
	ID             int      `json:"id"`
	ExpiryDate     string   `json:"expiry_date"`
	Nameservers    []string `json:"nameservers"`
	RegisteredDate string   `json:"registered_date"`
	Registrant     string   `json:"registrant"`
	Renew          bool     `json:"renew"`
	Services       Service  `json:"services"`
	Status         string
}

type Service struct {
	DNS       bool   `json:"dns"`
	Email     bool   `json:"email"`
	Registrar bool   `json:"registrar"`
	Webhotel  string `json:"webhotel"`
}

// DNSRecord JSON data structure.
type DNSRecord struct {
	Data string `json:"data"`
	Host string `json:"host"`
	ID   int    `json:"id"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}
