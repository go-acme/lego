package internal

type APIResponse struct {
	Message string `json:"message"`
	Result  int    `json:"result"`
}

// Record holds the Sonic API representation of a Domain Record.
type Record struct {
	UserID   string `json:"userid"`
	APIKey   string `json:"apikey"`
	Hostname string `json:"hostname"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	Type     string `json:"type"`
}
