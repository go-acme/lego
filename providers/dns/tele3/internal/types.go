package internal

type Operation struct {
	Key       string `json:"key,omitempty"`
	Secret    string `json:"secret,omitempty"`
	Operation string `json:"ope,omitempty"`
	Domain    string `json:"domain,omitempty"`
	Value     string `json:"value,omitempty"`
}
