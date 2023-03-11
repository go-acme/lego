package internal

type ZoneRecord struct {
	ID       string `json:"id,omitempty"`
	Domain   string `json:"domain"`
	Host     string `json:"host"`
	TTL      string `json:"ttl"`
	Priority string `json:"prio"`
	Type     string `json:"type"`
	Rdata    string `json:"rdata"`
	LastMod  string `json:"last_mod,omitempty"`
	Revoked  int    `json:"revoked,omitempty"`
	NewHost  string `json:"new_host,omitempty"`
}

type addRecordResponse struct {
	Msg    string     `json:"msg"`
	Tm     int        `json:"tm"`
	Data   ZoneRecord `json:"data"`
	Status int        `json:"status"`
}
