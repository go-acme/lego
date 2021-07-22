package internal

type zoneResponse struct {
	Name string `json:"name"`
}

type zoneRecord struct {
	TTL             int              `json:"ttl"`
	ResourceRecords []resourceRecord `json:"resource_records"`
}

type resourceRecord struct {
	Content []string `json:"content"`
}
