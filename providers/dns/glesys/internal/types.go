package internal

type addRecordRequest struct {
	DomainName string `json:"domainname"`
	Host       string `json:"host"`
	Type       string `json:"type"`
	Data       string `json:"data"`
	TTL        int    `json:"ttl,omitempty"`
}

type deleteRecordRequest struct {
	RecordID int `json:"recordid"`
}

type apiResponse struct {
	Response Response `json:"response"`
}

type Response struct {
	Status Status `json:"status"`
	Record Record `json:"record"`
}

type Status struct {
	Code int `json:"code"`
}

type Record struct {
	RecordID int `json:"recordid"`
}
