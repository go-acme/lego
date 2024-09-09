package internal

type Auth struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

type Token struct {
	Token   string `json:"token,omitempty"`
	Expires int    `json:"expires,omitempty"`
}

type Zone struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type ZoneDetails struct {
	Active bool     `json:"active,omitempty"`
	DNSSec bool     `json:"dnssec,omitempty"`
	Master string   `json:"master,omitempty"`
	Name   string   `json:"name,omitempty"`
	TSIG   *TSIGKey `json:"tsig,omitempty"`
	Type   string   `json:"type,omitempty"`
}

type TSIGKey struct {
	Algo   string `json:"algo,omitempty"`
	Secret string `json:"secret,omitempty"`
}

type Record struct {
	Name string `json:"name,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data,omitempty"`
}
