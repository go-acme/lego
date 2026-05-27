package internal

type APIError struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return a.Message
}

type Pager struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

type APIResponse[T any] struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       T           `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Record struct {
	ID       int    `json:"id,omitempty"`
	ZoneID   int    `json:"zone_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type RecordResponse struct {
	Record *Record `json:"record,omitempty"`
}

type ZonesResponse struct {
	Zones []Zone `json:"zones,omitempty"`
}

type Zone struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}
