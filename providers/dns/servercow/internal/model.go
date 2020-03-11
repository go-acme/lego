package internal

import "encoding/json"

// Record is the record representation.
type Record struct {
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl,omitempty"`

	Content Value `json:"content,omitempty"`
}

// Value is the value of a record.
// Allows to handle dynamic type (string and string array)
type Value []string

func (v Value) MarshalJSON() ([]byte, error) {
	if len(v) == 0 {
		return nil, nil
	}

	if len(v) == 1 {
		return json.Marshal(v[0])
	}

	content, err := json.Marshal([]string(v))
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (v *Value) UnmarshalJSON(b []byte) error {
	if b[0] == '[' {
		return json.Unmarshal(b, (*[]string)(v))
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*v = append(*v, s)
	return nil
}

// Message is the basic response representation.
// Can be an error.
type Message struct {
	Message  string `json:"message,omitempty"`
	ErrorMsg string `json:"error,omitempty"`
}

func (a Message) Error() string {
	return a.ErrorMsg
}
