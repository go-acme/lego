package internal

import (
	"fmt"
	"strings"
)

type APIError struct {
	Errors []Error `json:"errors"`
}

func (a *APIError) Error() string {
	var msg []string

	for _, e := range a.Errors {
		msg = append(msg, fmt.Sprintf("%s: %s", e.Code, e.Title))
	}

	return strings.Join(msg, ", ")
}

type Error struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

type Zone struct {
	Name          string        `json:"name,omitempty"`
	DomainConnect bool          `json:"domainConnect,omitempty"`
	Records       []Record      `json:"records"`
	URLForwards   []URLForward  `json:"urlForwards"`
	MailForwards  []MailForward `json:"mailForwards"`
	Report        *Report       `json:"report,omitempty"`
}

type Record struct {
	ID       int    `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Host     string `json:"host,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	RData    string `json:"rdata,omitempty"`
	Updated  *bool  `json:"updated"`
	Locked   *bool  `json:"locked"`
	IsDynDNS *bool  `json:"isDynDns"`
	Proxy    string `json:"proxy,omitempty"`
}

type URLForward struct {
	ID          int    `json:"id,omitempty"`
	ForwardType string `json:"forwardType,omitempty"`
	Host        string `json:"host,omitempty"`
	URL         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
	Description string `json:"description,omitempty"`
	Updated     *bool  `json:"updated,omitempty"`
}

type MailForward struct {
	ID          int    `json:"id,omitempty"`
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	Updated     *bool  `json:"updated,omitempty"`
}

type Report struct {
	IsValid           bool               `json:"isValid,omitempty"`
	RecordErrors      []RecordError      `json:"recordErrors,omitempty"`
	URLForwardErrors  []URLForwardError  `json:"urlForwardErrors,omitempty"`
	MailForwardErrors []MailForwardError `json:"mailForwardErrors,omitempty"`
	ZoneErrors        []ZoneError        `json:"zoneErrors,omitempty"`
}

func (r *Report) Error() string {
	var msg []string

	for _, e := range r.RecordErrors {
		msg = append(msg, e.Error())
	}

	for _, e := range r.URLForwardErrors {
		msg = append(msg, e.Error())
	}

	for _, e := range r.MailForwardErrors {
		msg = append(msg, e.Error())
	}

	for _, e := range r.ZoneErrors {
		msg = append(msg, e.Error())
	}

	return strings.Join(msg, ", ")
}

type RecordError struct {
	Messages []string `json:"messages,omitempty"`
	Record   Record   `json:"record"`
	Severity string   `json:"severity,omitempty"`
}

func (e *RecordError) Error() string {
	return fmt.Sprintf("record error (%s): %s", e.Severity, strings.Join(e.Messages, ", "))
}

type URLForwardError struct {
	Messages   []string   `json:"messages,omitempty"`
	URLForward URLForward `json:"urlForward"`
	Severity   string     `json:"severity,omitempty"`
}

func (e *URLForwardError) Error() string {
	return fmt.Sprintf("URL forward error (%s): %s", e.Severity, strings.Join(e.Messages, ", "))
}

type MailForwardError struct {
	Messages    []string    `json:"messages,omitempty"`
	MailForward MailForward `json:"mailForward"`
	Severity    string      `json:"severity,omitempty"`
}

func (e *MailForwardError) Error() string {
	return fmt.Sprintf("mail forward error (%s): %s", e.Severity, strings.Join(e.Messages, ", "))
}

type ZoneError struct {
	Message      string        `json:"message,omitempty"`
	Records      []Record      `json:"records,omitempty"`
	URLForwards  []URLForward  `json:"urlForwards,omitempty"`
	MailForwards []MailForward `json:"mailForwards,omitempty"`
	Severity     string        `json:"severity,omitempty"`
}

func (e *ZoneError) Error() string {
	return fmt.Sprintf("zone error (%s): %s", e.Severity, e.Message)
}
