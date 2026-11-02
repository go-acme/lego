package internal

import (
	"encoding/json"
	"strconv"
)

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
	EntityID string      `json:"entityID,omitempty"`
	Name     string      `json:"name,omitempty"`
	Type     string      `json:"type,omitempty"`
	TTL      StringOrInt `json:"ttl,omitempty"` // The value can be an integer (ex: `600`) or a string (`"-"`).
	Priority string      `json:"priority,omitempty"`
	Value    string      `json:"value,omitempty"`
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

type StringOrInt string

func (m *StringOrInt) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	raw := string(data)
	if data[0] == '"' {
		var err error

		raw, err = strconv.Unquote(string(data))
		if err != nil {
			return err
		}
	}

	*m = StringOrInt(raw)

	return nil
}
