package internal

type getZoneResponse struct {
	Name string `json:"name"`
}

type addRecordRequest struct {
	TTL             int              `json:"ttl"`
	ResourceRecords []resourceRecord `json:"resource_records"`
}

type resourceRecord struct {
	Content []string `json:"content"`
}
