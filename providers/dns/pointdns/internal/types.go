package internal

type APIError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	// TODO implement me
	panic("implement me")
}

type Record struct {
	ID     int64  `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Data   string `json:"data,omitempty"`
	TTL    int    `json:"ttl,omitempty"`
	Type   string `json:"record_type,omitempty"`
	AUX    string `json:"aux,omitempty"`
	ZoneID int64  `json:"zone_id,omitempty"`
}

type CreateZoneRecordResponse struct {
	ZoneRecord *Record `json:"zone_record"`
}
