package internal

type TXTRecord struct {
	Hostname    string `json:"hostName,omitempty"`
	Text        string `json:"text,omitempty"`
	TTL         int    `json:"ttl,omitempty"`
	PublishZone int    `json:"publishZone,omitempty"`
}
