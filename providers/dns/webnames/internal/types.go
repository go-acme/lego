package internal

import "encoding/json"

type APIResponse struct {
	Result  string          `json:"result"`
	Details json.RawMessage `json:"details"`
}
