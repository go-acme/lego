package internal

import "encoding/xml"

// kasAuthEnvelope a KAS authentication request envelope.
const kasAuthEnvelope = `
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
		<Body>
				<KasAuth xmlns="https://kasserver.com/">
						<Params>%s</Params>
				</KasAuth>
		</Body>
</Envelope>`

// KasAuthEnvelope a KAS envelope of the authentication response.
type KasAuthEnvelope struct {
	XMLName xml.Name    `xml:"Envelope"`
	Body    KasAuthBody `xml:"Body"`
}

type KasAuthBody struct {
	KasAuthResponse *KasResponse `xml:"KasAuthResponse"`
	Fault           *Fault       `xml:"Fault"`
}

// ---

type AuthRequest struct {
	Login                 string `json:"kas_login,omitempty"`
	AuthData              string `json:"kas_auth_data,omitempty"`
	AuthType              string `json:"kas_auth_type,omitempty"`
	SessionLifetime       int    `json:"session_lifetime,omitempty"`
	SessionUpdateLifetime string `json:"session_update_lifetime,omitempty"`
}
