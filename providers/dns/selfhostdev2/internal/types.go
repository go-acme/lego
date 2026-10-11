package internal

const (
	ActionAdd    = "present"
	ActionRemove = "cleanup"
)

type Payload struct {
	APIKey   string `json:"api_key"`
	Action   string `json:"action"`
	RecordID int64  `json:"record_id"`
	Content  string `json:"content"`
}
