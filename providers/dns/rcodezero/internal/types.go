package internal

import "fmt"

type UpdateRRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records"`
	TTL        int      `json:"ttl"`
}

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (a APIResponse) Error() string {
	return fmt.Sprintf("%s: %s", a.Status, a.Message)
}
