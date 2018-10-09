package dreamhost

const (
	cmdAddRecord    = "dns-add_record"
	cmdRemoveRecord = "dns-remove_record"
	defaultBaseURL  = "https://api.dreamhost.com"
)

// types for JSON responses

type apiResponse struct {
	Data   string `json:"data"`
	Result string `json:"result"`
}
