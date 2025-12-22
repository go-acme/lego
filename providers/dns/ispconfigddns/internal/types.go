package internal

type UpdateRecord struct {
	Action string `url:"action,omitempty"`
	Zone   string `url:"zone,omitempty"`
	Type   string `url:"type,omitempty"`
	Record string `url:"record,omitempty"`
	Data   string `url:"data,omitempty"`
}
