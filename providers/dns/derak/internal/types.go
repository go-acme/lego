package internal

import "time"

type GetRecordsParameters struct {
	DNSType string `url:"dnsType,omitempty"`
	Content string `url:"content,omitempty"`
}

type GetRecordsResponse struct {
	Data  []Record `json:"data"`
	Count int      `json:"count"`
}

type Record struct {
	Type    string `json:"type,omitempty"`
	Host    string `json:"host,omitempty"`
	Content string `json:"content,omitempty"`

	ID string `json:"recordId,omitempty"`

	TTL              int    `json:"ttl,omitempty"`
	Cloud            bool   `json:"cloud,omitempty"`
	Priority         int    `json:"priority,omitempty"`
	Service          string `json:"service,omitempty"`
	Protocol         string `json:"protocol,omitempty"`
	Weight           int    `json:"weight,omitempty"`
	Port             int    `json:"port,omitempty"`
	Advanced         bool   `json:"advanced,omitempty"`
	UpstreamPort     int    `json:"upstreamPort,omitempty"`
	UpstreamProtocol string `json:"upstreamProtocol,omitempty"`
	CustomSSLType    string `json:"customSSLType,omitempty"`
}

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Result  T    `json:"result"`
	Error   int  `json:"error"`
}

type Zone struct {
	ID               string    `json:"zoneId,omitempty"`
	Tags             []string  `json:"tags,omitempty"`
	ContextID        string    `json:"contextId,omitempty"`
	ContextType      string    `json:"contextType,omitempty"`
	HumanReadable    string    `json:"humanReadable,omitempty"`
	Serial           string    `json:"serial,omitempty"`
	CreationTime     int64     `json:"creationTime,omitempty"`
	CreationTimeDate time.Time `json:"creationTimeDate,omitzero"`
	Status           string    `json:"status,omitempty"`
	IsMoved          bool      `json:"is_moved,omitempty"`
	Paused           bool      `json:"paused,omitempty"`
	ServiceType      string    `json:"serviceType,omitempty"`
	Limbo            bool      `json:"limbo,omitempty"`
	TeamName         string    `json:"teamName,omitempty"`
	TeamID           string    `json:"teamId,omitempty"`
	MyTeam           bool      `json:"myTeam,omitempty"`
	RoleName         string    `json:"roleName,omitempty"`
	IsBoard          bool      `json:"isBoard,omitempty"`
	BoardRole        []string  `json:"boardRole,omitempty"`
}

func codeText(code int) string {
	switch code {
	case 1008:
		return "DNSValidationError"
	case 1003:
		return "ForbiddenError"
	case 1013:
		return "RateLimitExceeded"
	case 1021:
		return "RecordNotFoundError"
	default:
		return ""
	}
}
