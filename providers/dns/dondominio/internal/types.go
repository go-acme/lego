package internal

import "encoding/json"

type APIResponse struct {
	Success bool `json:"success"`

	Version string `json:"version"`
	Action  string `json:"action"`

	ErrorCode    int      `json:"errorCode"`
	ErrorCodeMsg string   `json:"errorCodeMsg"`
	Messages     []string `json:"messages"`

	Data json.RawMessage `json:"responseData,omitempty"`
}

type DNSResponse struct {
	DNS []Record `json:"dns"`
}

type Record struct {
	EntityID string `json:"entityID,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      string `json:"ttl,omitempty"`
	Priority string `json:"priority,omitempty"`
	Value    string `json:"value,omitempty"`
}

type CreateRecordRequest struct {
	ServiceName string `url:"serviceName,omitempty"`
	Name        string `url:"name,omitempty"`
	Type        string `url:"type,omitempty"`
	Value       string `url:"value,omitempty"`
	TTL         string `url:"ttl,omitempty"`
}

type DeleteRecordRequest struct {
	ServiceName string `url:"serviceName,omitempty"`
	EntityID    string `url:"entityID,omitempty"`
}
