package internal

import "fmt"

// https://api.mittwald.de/v2/docs/#/Domain/domain-list-domains

type Domain struct {
	Domain    string `json:"domain,omitempty"`
	DomainID  string `json:"domainId,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
}

// https://api.mittwald.de/v2/docs/#/Domain/dns-list-dns-zones

type DNSZone struct {
	ID        string     `json:"id,omitempty"`
	Domain    string     `json:"domain,omitempty"`
	RecordSet *RecordSet `json:"recordSet,omitempty"`
}

type RecordSet struct {
	TXT *TXTRecord `json:"txt"`
}

// https://api.mittwald.de/v2/docs/#/Domain/dns-create-dns-zone

type CreateDNSZoneRequest struct {
	Name         string `json:"name,omitempty"`
	ParentZoneID string `json:"parentZoneId,omitempty"`
}

type NewDNSZone struct {
	ID string `json:"id"`
}

// https://api.mittwald.de/v2/docs/#/Domain/dns-update-record-set

type TXTRecord struct {
	Settings Settings `json:"settings,omitempty"`
	Entries  []string `json:"entries,omitempty"`
}

type Settings struct {
	TTL TTL `json:"ttl"`
}

type TTL struct {
	Seconds int  `json:"seconds,omitempty"`
	Auto    bool `json:"auto,omitempty"`
}

// Error

type APIError struct {
	Type             string            `json:"type,omitempty"`
	Message          string            `json:"message,omitempty"`
	ValidationErrors []ValidationError `json:"validationErrors,omitempty"`
}

func (a APIError) Error() string {
	msg := fmt.Sprintf("%s: %s", a.Type, a.Message)

	if len(a.ValidationErrors) > 0 {
		for _, validationError := range a.ValidationErrors {
			msg += fmt.Sprintf(" [%s: %s (%s, %s)]",
				validationError.Type, validationError.Message, validationError.Path, validationError.Context.Format)
		}
	}

	return msg
}

type ValidationError struct {
	Message string                 `json:"message,omitempty"`
	Path    string                 `json:"path,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Context ValidationErrorContext `json:"context,omitempty"`
}

type ValidationErrorContext struct {
	Format string `json:"format,omitempty"`
}
