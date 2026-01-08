package internal

import "strings"

type APIResponse[T any] struct {
	Data     T         `json:"data,omitempty"`
	Page     *Page     `json:"page,omitempty"`
	Errors   Errors    `json:"errors,omitempty"`
	Warnings []Warning `json:"warnings,omitempty"`
}

type Errors []Error

func (e Errors) Error() string {
	if e == nil {
		return "API error"
	}

	var allMsg []string

	for _, item := range e {
		var msg []string

		if item.ID != "" {
			msg = append(msg, "id: "+item.ID)
		}

		if item.Code != "" {
			msg = append(msg, "code: "+item.Code)
		}

		if item.Title != "" {
			msg = append(msg, "title: "+item.Title)
		}

		if item.Detail != "" {
			msg = append(msg, "detail: "+item.Detail)
		}

		if item.Location != "" {
			msg = append(msg, "location: "+item.Location)
		}

		if item.Pointer != "" {
			msg = append(msg, "pointer: "+item.Pointer)
		}

		allMsg = append(allMsg, strings.Join(msg, ", "))
	}

	return strings.Join(allMsg, "; ")
}

type Error struct {
	ID       string `json:"id,omitempty"`
	Code     string `json:"code,omitempty"`
	Title    string `json:"title,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Location string `json:"location,omitempty"`
	Pointer  string `json:"pointer,omitempty"`
}

type Warning struct {
	Title  string `json:"title,omitempty"`
	Code   string `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type Page struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	Total  int `json:"total,omitempty"`
}

type RData struct {
	Value string `json:"value,omitempty"`
	Label string `json:"label,omitempty"`
}
