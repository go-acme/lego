package rimuhosting

import "encoding/xml"

type ActionParameter struct {
	Action   string `url:"action,omitempty"`
	Name     string `url:"name,omitempty"`
	Type     string `url:"type,omitempty"`
	Value    string `url:"value,omitempty"`
	TTL      int    `url:"ttl,omitempty"`
	Priority int    `url:"prio,omitempty"`
}

type actionParameter struct {
	ActionParameter

	APIKey string `url:"api_key,omitempty"`
}

type multiActionParameter struct {
	APIKey string `url:"api_key,omitempty"`

	Action   []string `url:"action,brackets,numbered,omitempty"`
	Name     []string `url:"name,brackets,numbered,omitempty"`
	Type     []string `url:"type,brackets,numbered,omitempty"`
	Value    []string `url:"value,brackets,numbered,omitempty"`
	TTL      []int    `url:"ttl,brackets,numbered,omitempty"`
	Priority []int    `url:"prio,brackets,numbered,omitempty"`
}

type APIError struct {
	XMLName xml.Name `xml:"error"`
	Text    string   `xml:",chardata"`
}

func (a APIError) Error() string {
	return a.Text
}

type DNSAPIResult struct {
	XMLName      xml.Name     `xml:"dnsapi_result"`
	IsOk         string       `xml:"is_ok"`
	ResultCounts ResultCounts `xml:"result_counts"`
	Actions      Actions      `xml:"actions"`
}

type ResultCounts struct {
	Added     string `xml:"added,attr"`
	Changed   string `xml:"changed,attr"`
	Unchanged string `xml:"unchanged,attr"`
	Deleted   string `xml:"deleted,attr"`
}

type Actions struct {
	Action Action `xml:"action"`
}

type Action struct {
	Action  string   `xml:"action,attr"`
	Host    string   `xml:"host,attr"`
	Type    string   `xml:"type,attr"`
	Records []Record `xml:"record"`
}

type Record struct {
	Name     string `xml:"name,attr"`
	Type     string `xml:"type,attr"`
	Content  string `xml:"content,attr"`
	TTL      string `xml:"ttl,attr"`
	Priority string `xml:"prio,attr"`
}
