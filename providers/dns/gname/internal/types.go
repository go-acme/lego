package internal

import "encoding/json"

type Record struct {
	RecordID    int    `url:"jxid,omitempty"`
	Domain      string `url:"ym,omitempty"`
	RecordType  string `url:"lx,omitempty"`
	HostRecord  string `url:"zj,omitempty"`
	RecordValue string `url:"jlz,omitempty"`
	MX          int    `url:"mx,omitempty"`
	TTL         int    `url:"ttl,omitempty"`
	Router      string `url:"xl,omitempty"`
}

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}
