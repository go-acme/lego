package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Status   string            `json:"status"`
	ErrorMsg string            `json:"error"`
	Code     int               `json:"code"`
	Context  map[string]string `json:"context"`
}

func (a *APIError) Error() string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("status: %s, error: %s", a.Status, a.ErrorMsg))

	if a.Code != 0 {
		msg.WriteString(fmt.Sprintf(", code: %d", a.Code))
	}

	if len(a.Context) != 0 {
		for k, v := range a.Context {
			msg.WriteString(fmt.Sprintf(", %s: %s", k, v))
		}
	}

	return msg.String()
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Auth struct {
	Successful bool `json:"successful"`
}

type UserInfo struct {
	Username      string       `json:"username"`
	Authenticated bool         `json:"authenticated"`
	Permissions   []Permission `json:"permissions"`
}

type Permission struct {
	Methods []string `json:"methods"`
	Path    string   `json:"path"`
}

type Zones struct {
	Zones []Zone `json:"zones"`
}

type Zone struct {
	Name           string          `json:"name"`
	HandlerConfigs []HandlerConfig `json:"handlerConfigs"`
	DefaultTTL     int             `json:"defaultTTL"`
	Authoritative  bool            `json:"authoritative"`
	Hook           string          `json:"hook"`
	RecordCount    int             `json:"recordCount"`
}

type HandlerConfig struct {
	Type     string   `json:"type"`
	CacheTTL int      `json:"cache_ttl,omitempty"`
	To       []string `json:"to,omitempty"`
}

type Record struct {
	Data         string `json:"data,omitempty"`
	Fqdn         string `json:"fqdn,omitempty"`
	Hostname     string `json:"hostname,omitempty"`
	MxPreference int    `json:"mxPreference,omitempty"`
	SrvPort      int    `json:"srvPort,omitempty"`
	SrvPriority  int    `json:"srvPriority,omitempty"`
	SrvWeight    int    `json:"srvWeight,omitempty"`
	Type         string `json:"type,omitempty"`
	UID          string `json:"uid,omitempty"`
}
