package internal

type Record struct {
	Hostname string `url:"hostname"`
	Type     string `url:"type"`
	TTL      int    `url:"ttl"`
	Value    string `url:"value"`
}
