package internal

// Zone represents a DNScale DNS zone.
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ZonesData wraps the zone list in a paginated payload.
type ZonesData struct {
	Zones []Zone `json:"zones"`
}

// ZonesResponse is the API response for listing zones.
// The DNScale API wraps responses as {"status":"success","data":{...}}.
type ZonesResponse struct {
	Status string    `json:"status"`
	Data   ZonesData `json:"data"`
}

// RecordRequest is the request body for creating a DNS record.
type RecordRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// APIErrorDetails holds the nested error fields.
type APIErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIError is the error response from the DNScale API.
// Shape: {"status":"error","error":{"code":"...","message":"..."}}.
type APIError struct {
	Status string          `json:"status"`
	Error  APIErrorDetails `json:"error"`
}
