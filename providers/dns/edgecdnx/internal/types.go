package internal

import "fmt"

const RoutingPolicySimple = "Simple"

type APIError struct {
	Message    string `json:"error"`
	StatusCode int    `json:"-"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("[status code %d]: %s", a.StatusCode, a.Message)
}

type Record ResponseEnvelope[RecordSpec]

type Zone ResponseEnvelope[ZoneSpec]

type ResponseEnvelope[T any] struct {
	Metadata Metadata `json:"metadata"`
	Spec     T        `json:"spec"`
}

type Metadata struct {
	Name            string `json:"name"`
	ResourceVersion string `json:"resourceVersion"`
}

type ZoneSpec struct {
	Zone string `json:"zone"`
}

type RecordSpec struct {
	ResourceVersion string   `json:"resourceVersion,omitempty"`
	DNSName         string   `json:"dnsName,omitempty"`
	TTL             int      `json:"recordTTL,omitempty"`
	Type            string   `json:"recordType,omitempty"`
	RoutingPolicy   string   `json:"routingPolicy,omitempty"`
	Targets         []string `json:"targets,omitempty"`
}
