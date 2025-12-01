package internal

import (
	"fmt"
	"strings"
)

type APIResponse[T any] struct {
	STID     string            `json:"stid"`
	CTID     string            `json:"ctid"`
	Messages []ResponseMessage `json:"messages"`
	Status   *ResponseStatus   `json:"status"`
	Object   *ResponseObject   `json:"object"`
	Data     T                 `json:"data"`
}

type APIError APIResponse[any]

func (a *APIError) Error() string {
	var parts []string

	if a.STID != "" {
		parts = append(parts, fmt.Sprintf("STID: %s", a.STID))
	}

	if a.CTID != "" {
		parts = append(parts, fmt.Sprintf("CTID: %s", a.CTID))
	}

	if a.Status != nil {
		parts = append(parts, "status: "+a.Status.String())
	}

	for _, message := range a.Messages {
		parts = append(parts, "message: "+message.String())
	}

	if a.Object != nil {
		parts = append(parts, "object: "+a.Object.String())
	}

	return strings.Join(parts, ", ")
}

type DataZoneResponse APIResponse[[]Zone]

type ResponseMessage struct {
	Text     string          `json:"text"`
	Code     string          `json:"code"`
	Status   string          `json:"status"`
	Messages []string        `json:"messages"`
	Objects  []GenericObject `json:"objects"`
}

func (r ResponseMessage) String() string {
	var parts []string

	if r.Code != "" {
		parts = append(parts, "code: "+r.Code)
	}

	if r.Text != "" {
		parts = append(parts, "text: "+r.Text)
	}

	if r.Status != "" {
		parts = append(parts, "status: "+r.Status)
	}

	if len(r.Messages) > 0 {
		parts = append(parts, "messages: "+strings.Join(r.Messages, ";"))
	}

	for _, object := range r.Objects {
		parts = append(parts, fmt.Sprintf("object: %s", object))
	}

	return strings.Join(parts, ", ")
}

type GenericObject struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (g GenericObject) String() string {
	return g.Type + ": " + g.Value
}

type ResponseStatus struct {
	Code string `json:"code"`
	Text string `json:"text"`
	Type string `json:"type"` // SUCCESS, ERROR, NOTIFY, NOTICE, NICCOM_NOTIFY
}

func (r ResponseStatus) String() string {
	return fmt.Sprintf("code: %s, text: %s, type: %s", r.Code, r.Text, r.Type)
}

type ResponseObject struct {
	Type    string              `json:"type"`
	Value   string              `json:"value"`
	Summary int32               `json:"summary"`
	Data    *ResponseObjectData `json:"data"`
}

func (r ResponseObject) String() string {
	var parts []string

	if r.Type != "" {
		parts = append(parts, fmt.Sprintf("type: %s", r.Type))
	}

	if r.Value != "" {
		parts = append(parts, fmt.Sprintf("value: %s", r.Value))
	}

	if r.Summary != 0 {
		parts = append(parts, fmt.Sprintf("summary: %d", r.Summary))
	}

	if r.Data != nil {
		parts = append(parts, fmt.Sprintf("data: %s", r.Data.Description))
	}

	return strings.Join(parts, ", ")
}

type ResponseObjectData struct {
	Description string `json:"description"`
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
	Name              string           `json:"origin"`
	ResourceRecords   []ResourceRecord `json:"resourceRecords"`
	Action            string           `json:"action"`
	VirtualNameServer string           `json:"virtualNameServer"`
}

// ZoneStream body of the requests.
// https://github.com/InterNetX/domainrobot-api/blob/bdc8fe92a2f32fcbdb29e30bf6006ab446f81223/src/domainrobot.json#L35914-L35932
type ZoneStream struct {
	Adds    []*ResourceRecord `json:"adds"`
	Removes []*ResourceRecord `json:"rems"`
}
