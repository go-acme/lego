package internal

const codeOk = 1000

const (
	CommandPing            = "ping"
	CommandDNSDomainCommit = "dns-domain-commit"
	CommandDNSRowsList     = "dns-rows-list"
	CommandDNSRowDelete    = "dns-row-delete"
	CommandDNSRowAdd       = "dns-row-add"
	CommandDNSRowUpdate    = "dns-row-update"
)

type Response interface {
	GetCode() int
	GetResult() string
}

type ResponseEnvelope[D any] struct {
	Response ResponsePayload[D] `json:"response"`
}

func (a ResponseEnvelope[D]) GetCode() int {
	return a.Response.Code
}

func (a ResponseEnvelope[D]) GetResult() string {
	return a.Response.Result
}

type ResponsePayload[D any] struct {
	Code            int    `json:"code"`
	Result          string `json:"result"`
	Timestamp       int    `json:"timestamp"`
	ServerRequestID string `json:"svTRID"`
	Command         string `json:"command"`
	Data            D      `json:"data"`
}

type Rows struct {
	Records []Record `json:"row"`
}

type Record struct {
	ID   string `json:"ID"`
	Name string `json:"name"`
	TTL  string `json:"ttl"`
	Type string `json:"rdtype"`
	Data string `json:"rdata"`
}

type RequestEnvelope[D any] struct {
	Request RequestPayload[D] `json:"request"`
}

type RequestPayload[D any] struct {
	User    string `json:"user,omitempty"`
	Auth    string `json:"auth,omitempty"`
	Command string `json:"command,omitempty"`
	Data    D      `json:"data,omitempty"`
}

type RecordRequest struct {
	ID     string `json:"row_id,omitempty"`
	Domain string `json:"domain,omitempty"`
	Name   string `json:"name,omitempty"`
	TTL    int    `json:"ttl,omitempty"`
	Type   string `json:"type,omitempty"`
	Data   string `json:"rdata"`
}
