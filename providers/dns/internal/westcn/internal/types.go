package internal

import "fmt"

type APIResponse[T any] struct {
	Result    int    `json:"result,omitempty"`
	ClientID  string `json:"clientid,omitempty"`
	Message   string `json:"msg,omitempty"`
	ErrorCode int    `json:"errcode,omitempty"`
	Data      T      `json:"data,omitempty"`
}

func (a APIResponse[T]) Error() string {
	return fmt.Sprintf("%d: %s (%d)", a.ErrorCode, a.Message, a.Result)
}

type Record struct {
	Domain   string `url:"domain,omitempty"`
	Host     string `url:"host,omitempty"`
	Type     string `url:"type,omitempty"`
	Value    string `url:"value,omitempty"`
	TTL      int    `url:"ttl,omitempty"` // 60~86400 seconds
	Priority int    `url:"level,omitempty"`
}

type RecordID struct {
	ID int `json:"id,omitempty"`
}
