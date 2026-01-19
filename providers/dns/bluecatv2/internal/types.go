package internal

import (
	"fmt"
	"time"
)

// Quick deployment states.
//
//nolint:misspell // US vs UK
const (
	QDStatePending               = "PENDING"
	QDStateQueued                = "QUEUED"
	QDStateRunning               = "RUNNING"
	QDStateCancelled             = "CANCELLED"
	QDStateCancelling            = "CANCELLING"
	QDStateCompleted             = "COMPLETED"
	QDStateCompletedWithErrors   = "COMPLETED_WITH_ERRORS"
	QDStateCompletedWithWarnings = "COMPLETED_WITH_WARNINGS"
	QDStateFailed                = "FAILED"
	QDStateUnknown               = "UNKNOWN"
)

// APIError represents an error.
// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Errors/9.6.0
type APIError struct {
	Status  int    `json:"status"`
	Reason  string `json:"reason"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s: %s: %s", a.Status, a.Reason, a.Code, a.Message)
}

// CommonResource represents the common resource fields.
// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Resources/9.6.0
type CommonResource struct {
	ID   int64  `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

// Collection represents a collection of resources.
// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Collections/9.6.0
type Collection[T any] struct {
	Count      int64 `json:"count"`
	TotalCount int64 `json:"totalCount"`
	Data       []T   `json:"data"`
}

type CollectionOptions struct {
	// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Fields/9.6.0
	Fields string `url:"fields,omitempty"`

	// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Pagination/9.6.0
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`

	// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Filter/9.6.0
	Filter string `url:"filter,omitempty"`

	// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Ordering/9.6.0
	OrderBy string `url:"orderBy,omitempty"`

	// Should return or not the total number of resources matching the query.
	Total bool `url:"total,omitempty"`
}

type RecordTXT struct {
	CommonResource

	TTL          int    `json:"ttl,omitempty"`
	AbsoluteName string `json:"absoluteName,omitempty"`
	Comment      string `json:"comment,omitempty"`
	Dynamic      bool   `json:"dynamic,omitempty"`
	RecordType   string `json:"recordType,omitempty"`
	Text         string `json:"text,omitempty"`
}

type ZoneResource struct {
	CommonResource

	AbsoluteName string `json:"absoluteName,omitempty"`
}

type QuickDeployment struct {
	CommonResource

	State              string    `json:"state,omitempty"`
	Status             string    `json:"status,omitempty"`
	Message            string    `json:"message,omitempty"`
	PercentComplete    int       `json:"percentComplete,omitempty"`
	CreationDateTime   time.Time `json:"creationDateTime,omitzero"`
	StartDateTime      time.Time `json:"startDateTime,omitzero"`
	CompletionDateTime time.Time `json:"completionDateTime,omitzero"`
	Method             string    `json:"method,omitempty"`
}

// LoginInfo represents the login information.
// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Creating-an-API-session/9.6.0
type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Session represents the session.
// https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Creating-an-API-session/9.6.0
type Session struct {
	ID                             int       `json:"id"`
	Type                           string    `json:"type"`
	APIToken                       string    `json:"apiToken"`
	APITokenExpirationDateTime     time.Time `json:"apiTokenExpirationDateTime"`
	BasicAuthenticationCredentials string    `json:"basicAuthenticationCredentials"`
	RemoteAddress                  string    `json:"remoteAddress"`
	ReadOnly                       bool      `json:"readOnly"`
	LoginDateTime                  time.Time `json:"loginDateTime"`
	LogoutDateTime                 time.Time `json:"logoutDateTime"`
	State                          string    `json:"state"`
	Response                       string    `json:"response"`
}
