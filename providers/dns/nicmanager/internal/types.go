package internal

type Record struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type Zone struct {
	Name    string   `json:"name"`
	Active  bool     `json:"active"`
	Records []Record `json:"records"`
}

type RecordCreateUpdate struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
	Type  string `json:"type"`
}
