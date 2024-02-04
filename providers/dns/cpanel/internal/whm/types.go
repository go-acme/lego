package whm

import (
	"fmt"

	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/shared"
)

type Metadata struct {
	Command string `json:"command,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Result  int    `json:"result,omitempty"`
	Version int    `json:"version,omitempty"`
}

type ZoneData struct {
	Payload []shared.ZoneRecord `json:"payload,omitempty"`
}

type APIResponse[T any] struct {
	Data     T        `json:"data,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

func toError(m Metadata) error {
	return fmt.Errorf("%s failure: %s", m.Command, m.Reason)
}
