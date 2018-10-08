package dreamhost

const defaultBaseURL = "https://api.dreamhost.com"

// types for JSON responses

type responseStruct struct {
	Data   string `json:"data"`
	Result string `json:"result"`
}
