package internal

type AddRecordResponse struct {
	Data struct {
		ID int `json:"id"`
	} `json:"data"`
	Meta struct {
		Location string `json:"location"`
	}
}

type Record struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type APIError struct {
	Message string `json:"message"`
}

func (a APIError) Error() string {
	return a.Message
}
