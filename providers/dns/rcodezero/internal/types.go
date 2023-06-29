package internal

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type UpdateRRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records"`
	TTL        int      `json:"ttl"`
}

type apiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (a apiResponse) Error() string {
	return a.Message
}
