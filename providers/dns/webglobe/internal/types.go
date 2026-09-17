package internal

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	Message string          `json:"message,omitempty"`
	Code    int             `json:"code,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.Code, a.Message)
}

type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Status  string    `json:"status"`
	Error   *APIError `json:"error"`
	Data    T         `json:"data"`
}

type RecordResponse struct {
	APIResponse[*Record]

	NewID int64 `json:"new_id"`
}

type Domain struct {
	DomainID    int64  `json:"domain_id"`
	Domain      string `json:"domain"`
	Status      string `json:"status"`
	AltName     string `json:"alt_name"`
	ObjectType  string `json:"object_type"`
	HostingOnly int    `json:"hosting_only"`
	State       string `json:"state"`
	TLD         string `json:"tld"`
	RegStatus   string `json:"reg_status"`
}

type Record struct {
	ID   int64  `json:"id,omitempty"`
	Zone int64  `json:"zone,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Data string `json:"data,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
	AUX  int    `json:"aux,omitempty"`
}

type Auth struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	OTP      string `json:"otp,omitempty"`
}

type Token struct {
	CustomerLoginID   int64  `json:"IDcustlogin"`
	CustomerID        int64  `json:"IDcustomer"`
	CustomerLoginName string `json:"custlogin_name"`
	SessionID         string `json:"session_id"`
	Token             string `json:"token"`
	TokenType         string `json:"token_type"`
	ExpiresIn         int64  `json:"expires_in"`
}
