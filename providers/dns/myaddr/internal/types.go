package internal

type ACMEChallenge struct {
	Key  string `json:"key"`
	Data string `json:"acme_challenge"`
}
