package active24

import "fmt"

type APIError struct {
	// v2 error
	Type   string `json:"type,omitempty"`
	Status int    `json:"status,omitempty"`
	Title  string `json:"title,omitempty"`

	// v1 error
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

func (a *APIError) Error() string {
	if a.Message != "" {
		return fmt.Sprintf("%d: %s", a.Code, a.Message)
	}

	return fmt.Sprintf("%d: %s: %s", a.Status, a.Type, a.Title)
}

type APIResponse struct {
	Data []Record `json:"data"`
}

type Record struct {
	ID       int    `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Port     int    `json:"port,omitempty"`
	Weight   int    `json:"weight,omitempty"`
}

type OldAPIResponse struct {
	Items []Service `json:"items"`
}

type Service struct {
	ID          int     `json:"id,omitempty"`
	ServiceName string  `json:"serviceName,omitempty"`
	Status      string  `json:"status,omitempty"`
	Name        string  `json:"name,omitempty"`
	CreateTime  int     `json:"createTime,omitempty"`
	ExpireTime  int     `json:"expireTime,omitempty"`
	Price       float64 `json:"price,omitempty"`
	AutoExtend  bool    `json:"autoExtend,omitempty"`
}

type RecordFilter struct {
	Name     string   `json:"name,omitempty"`
	Type     []string `json:"type,omitempty"`
	Content  string   `json:"content,omitempty"`
	TTL      int      `json:"ttl,omitempty"`
	Note     string   `json:"note,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Port     int      `json:"port,omitempty"`
	Weight   int      `json:"weight,omitempty"`
	Flags    int      `json:"flags,omitempty"`
	Tag      []string `json:"tag,omitempty"`
}
