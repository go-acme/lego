package internal

type apiResponse[T any] struct {
	Msg    string `json:"msg"`
	Status int    `json:"status"`
	Tm     int    `json:"tm"`
	Data   T      `json:"data"`
	Count  int    `json:"count"`
	Total  int    `json:"total"`
	Start  int    `json:"start"`
	Max    int    `json:"max"`
}

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
