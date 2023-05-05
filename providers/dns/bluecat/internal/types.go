package internal

// Entity JSON body for Bluecat entity requests.
type Entity struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}

// EntityResponse JSON body for Bluecat entity responses.
type EntityResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}
