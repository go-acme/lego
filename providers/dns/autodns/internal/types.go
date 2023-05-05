package internal

type ResponseMessage struct {
	Text     string   `json:"text"`
	Messages []string `json:"messages"`
	Objects  []string `json:"objects"`
	Code     string   `json:"code"`
	Status   string   `json:"status"`
}

type ResponseStatus struct {
	Code string `json:"code"`
	Text string `json:"text"`
	Type string `json:"type"`
}

type ResponseObject struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	Summary int32  `json:"summary"`
	Data    string
}

type DataZoneResponse struct {
	STID     string             `json:"stid"`
	CTID     string             `json:"ctid"`
	Messages []*ResponseMessage `json:"messages"`
	Status   *ResponseStatus    `json:"status"`
	Object   any                `json:"object"`
	Data     []*Zone            `json:"data"`
}

// ResourceRecord holds a resource record.
// https://help.internetx.com/display/APIXMLEN/Resource+Record+Object
type ResourceRecord struct {
	Name  string `json:"name"`
	TTL   int64  `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Pref  int32  `json:"pref,omitempty"`
}

// Zone is an autodns zone record with all for us relevant fields.
// https://help.internetx.com/display/APIXMLEN/Zone+Object
type Zone struct {
	Name              string            `json:"origin"`
	ResourceRecords   []*ResourceRecord `json:"resourceRecords"`
	Action            string            `json:"action"`
	VirtualNameServer string            `json:"virtualNameServer"`
}

// ZoneStream body of the requests.
// https://github.com/InterNetX/domainrobot-api/blob/bdc8fe92a2f32fcbdb29e30bf6006ab446f81223/src/domainrobot.json#L35914-L35932
type ZoneStream struct {
	Adds    []*ResourceRecord `json:"adds"`
	Removes []*ResourceRecord `json:"rems"`
}
