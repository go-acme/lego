package internal

import "encoding/xml"

// kasAPIEnvelope a KAS API request envelope.
const kasAPIEnvelope = `
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Body>
        <KasApi xmlns="https://kasserver.com/">
            <Params>%s</Params>
        </KasApi>
    </Body>
</Envelope>`

// KasAPIResponseEnvelope a KAS envelope of the API response.
type KasAPIResponseEnvelope struct {
	XMLName xml.Name   `xml:"Envelope"`
	Body    KasAPIBody `xml:"Body"`
}

type KasAPIBody struct {
	KasAPIResponse *KasResponse `xml:"KasApiResponse"`
	Fault          *Fault       `xml:"Fault"`
}

// ---

type KasRequest struct {
	// Login the relevant KAS login.
	Login string `json:"kas_login,omitempty"`
	// AuthType the authentication type.
	AuthType string `json:"kas_auth_type,omitempty"`
	// AuthData the authentication data.
	AuthData string `json:"kas_auth_data,omitempty"`
	// Action API function.
	Action string `json:"kas_action,omitempty"`
	// RequestParams Parameters to the API function.
	RequestParams any `json:"KasRequestParams,omitempty"`
}

type DNSRequest struct {
	// ZoneHost the zone in question (must be a FQDN).
	ZoneHost string `json:"zone_host"`
	// RecordType the TYPE of the resource record (MX, A, AAAA etc.).
	RecordType string `json:"record_type"`
	// RecordName the NAME of the resource record.
	RecordName string `json:"record_name"`
	// RecordData the DATA of the resource record.
	RecordData string `json:"record_data"`
	// RecordAux the AUX of the resource record.
	RecordAux int `json:"record_aux"`
}

// ---

type GetDNSSettingsAPIResponse struct {
	Response GetDNSSettingsResponse `json:"Response"  mapstructure:"Response"`
}

type GetDNSSettingsResponse struct {
	KasFloodDelay float64      `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    []ReturnInfo `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string       `json:"ReturnString"`
}

type ReturnInfo struct {
	ID         any    `json:"record_id,omitempty" mapstructure:"record_id"`
	Zone       string `json:"record_zone,omitempty" mapstructure:"record_zone"`
	Name       string `json:"record_name,omitempty" mapstructure:"record_name"`
	Type       string `json:"record_type,omitempty" mapstructure:"record_type"`
	Data       string `json:"record_data,omitempty" mapstructure:"record_data"`
	Changeable string `json:"record_changeable,omitempty" mapstructure:"record_changeable"`
	Aux        int    `json:"record_aux,omitempty" mapstructure:"record_aux"`
}

type AddDNSSettingsAPIResponse struct {
	Response AddDNSSettingsResponse `json:"Response" mapstructure:"Response"`
}

type AddDNSSettingsResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay" mapstructure:"KasFloodDelay"`
	ReturnInfo    string  `json:"ReturnInfo" mapstructure:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString" mapstructure:"ReturnString"`
}

type DeleteDNSSettingsAPIResponse struct {
	Response DeleteDNSSettingsResponse `json:"Response"`
}

type DeleteDNSSettingsResponse struct {
	KasFloodDelay float64 `json:"KasFloodDelay"`
	ReturnInfo    bool    `json:"ReturnInfo"`
	ReturnString  string  `json:"ReturnString"`
}
