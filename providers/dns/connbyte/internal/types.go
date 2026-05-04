package internal

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

type APIError struct {
	Status Status              `json:"status"`
	Field  []map[string]string `json:"field"`
}

func (a APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: %s %s (%s %d %s)", a.Status.Code, a.Status.Title, a.Status.Message, a.Status.ID, a.Status.Time, a.Status.Path)

	fields := make(map[string]string)

	for _, m := range a.Field {
		for k, v := range m {
			text, ok := fields[k]
			if ok {
				fields[k] = text + " " + v
			} else {
				fields[k] = v
			}
		}
	}

	for _, k := range slices.Sorted(maps.Keys(fields)) {
		_, _ = fmt.Fprintf(msg, " [%s: %s]", k, fields[k])
	}

	return msg.String()
}

type APIResponse[T any] struct {
	Status Status              `json:"status"`
	Data   T                   `json:"data,omitempty"`
	Field  []map[string]string `json:"field"`
}

type Status struct {
	Code            int    `json:"code"`
	Title           string `json:"title"`
	Message         string `json:"message"`
	ID              string `json:"id"`
	TitleOriginal   string `json:"titleOriginal"`
	MessageOriginal string `json:"messageOriginal"`
	LogID           string `json:"logId"`
	Path            string `json:"path"`
	Time            int    `json:"time"`
	Version         string `json:"version"`
}

type ZoneListResponse struct {
	Zone Zone `json:"zone"`
}

type RecordCreateResponse struct {
	ID     int64  `json:"id"`
	Record string `json:"record"`
}

type Zone struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	Zone      string    `json:"zone"`
	CreatedAt time.Time `json:"createdAt,omitzero"`
	UpdatedAt time.Time `json:"updatedAt,omitzero"`
}

type Record struct {
	ZoneID   string `json:"zoneId"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority"`
	Disabled bool   `json:"disabled"`
}
