package internal

type Record struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"prio,omitempty"`
}

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Data    []T  `json:"data"`
}

type APIError struct {
	ErrorMsg string `json:"error"`
	Errors   Error  `json:"errors"`
}

func (e APIError) Error() string {
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}

	return e.Errors.Error()
}

type Error struct {
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}
