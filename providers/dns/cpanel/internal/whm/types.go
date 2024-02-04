package whm

import (
	"fmt"

	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/shared"
)

type Metadata struct {
	Command string `json:"command"`
	Reason  string `json:"reason"`
	Result  int    `json:"result"`
	Version int    `json:"version"`
}

type ZoneData struct {
	Payload []shared.ZoneRecord `json:"payload"`
}

type APIResponse[T any] struct {
	Data     T        `json:"data"`
	Metadata Metadata `json:"metadata"`
}

func toError(m Metadata) error {
	return fmt.Errorf("%s failure: %s", m.Command, m.Reason)
}
