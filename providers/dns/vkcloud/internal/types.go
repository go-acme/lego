package internal

type DNSZone struct {
	UUID          string `json:"uuid,omitempty"`
	Tenant        string `json:"tenant,omitempty"`
	SoaPrimaryDNS string `json:"soa_primary_dns,omitempty"`
	SoaAdminEmail string `json:"soa_admin_email,omitempty"`
	SoaSerial     int    `json:"soa_serial,omitempty"`
	SoaRefresh    int    `json:"soa_refresh,omitempty"`
	SoaRetry      int    `json:"soa_retry,omitempty"`
	SoaExpire     int    `json:"soa_expire,omitempty"`
	SoaTTL        int    `json:"soa_ttl,omitempty"`
	Zone          string `json:"zone,omitempty"`
	Status        string `json:"status,omitempty"`
}

type DNSTXTRecord struct {
	UUID    string `json:"uuid,omitempty"`
	Name    string `json:"name,omitempty"`
	DNS     string `json:"dns,omitempty"`
	Content string `json:"content,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
}
