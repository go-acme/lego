package whm

import (
	"fmt"

	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/shared"
)

type APIResponse[T any] struct {
	Metadata Metadata `json:"metadata"`
	Data     T        `json:"data,omitempty"`
}

type Metadata struct {
	Command string `json:"command,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Result  int    `json:"result,omitempty"`
	Version int    `json:"version,omitempty"`
}

type ZoneData struct {
	Payload []shared.ZoneRecord `json:"payload,omitempty"`
}

func toError(m Metadata) error {
	return fmt.Errorf("%s error(%d): %s", m.Command, m.Result, m.Reason)
}
