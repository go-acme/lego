package internal

type createTXTRequest struct {
	Records []createTXTRecord `json:"records"`
}

type createTXTRecord struct {
	Host string `json:"host"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
	Data string `json:"data"`
}

type createTXTResponse struct {
	Added   int    `json:"records_added"`
	Removed int    `json:"records_removed"`
	Message string `json:"message"`
}

type deleteTXTResponse struct {
	Removed int    `json:"records_removed"`
	Message string `json:"message"`
}
