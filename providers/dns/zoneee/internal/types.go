package internal

type TXTRecord struct {
	// Identifier (identificator)
	ID string `json:"id,omitempty"`
	// Hostname
	Name string `json:"name"`
	// TXT content value
	Destination string `json:"destination"`
	// Can this record be deleted
	Delete bool `json:"delete,omitempty"`
	// Can this record be modified
	Modify bool `json:"modify,omitempty"`
	// API url to get this entity
	ResourceURL string `json:"resource_url,omitempty"`
}
