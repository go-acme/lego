package internal

import "encoding/json"

type APIError struct{}

func (a *APIError) Error() string {
	// TODO implement me
	panic("implement me")
}

type APIResponse struct {
	Code    int             `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type APIRequest[T any] struct {
	APIKey    string `json:"apiKey"`
	Timestamp string `json:"timestamp"`
	Hash      string `json:"hash"`
}

type RecordRequest struct {
	DomainID int64  `json:"domainID,omitempty" url:"domainID,omitempty"`
	Type     string `json:"type,omitempty" url:"type,omitempty"`
	ViewID   int64  `json:"viewID,omitempty" url:"viewID,omitempty"`
	Host     string `json:"host,omitempty" url:"host,omitempty"`
	Value    string `json:"value,omitempty" url:"value,omitempty"`
	TTL      int    `json:"ttl,omitempty" url:"ttl,omitempty"`
	MX       int    `json:"mx,omitempty" url:"mx,omitempty"`
	Remark   string `json:"remark,omitempty" url:"remark,omitempty"`
}

type RecordData struct {
	RecordRequest

	RecordID int64  `json:"recordID,omitempty"`
	State    int64  `json:"state,omitempty"`
	Record   string `json:"record,omitempty"`
}

type DomainRequest struct {
	GroupID int64 `json:"groupID,omitempty" url:"groupID,omitempty"`

	Page     int `json:"page,omitempty" url:"page,omitempty"`
	PageSize int `json:"pageSize,omitempty" url:"pageSize,omitempty"`
}

type DomainData struct {
	Data []Domain `json:"data,omitempty"`

	Page      int `json:"page,omitempty"`
	PageSize  int `json:"pageSize,omitempty"`
	PageCount int `json:"pageCount,omitempty"`
}

type Domain struct {
	GroupID        int64  `json:"groupID,omitempty"`
	DomainID       int64  `json:"domainID,omitempty"`
	Domain         string `json:"domains,omitempty"`
	State          int    `json:"state,omitempty"`
	UserLockState  int    `json:"userLock,omitempty"`
	AdminLockState int    `json:"adminLock,omitempty"`
}
