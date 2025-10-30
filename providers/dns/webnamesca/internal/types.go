package internal

import (
	"encoding/json"
	"fmt"
	"time"
)

type APIError struct {
	ErrorMessage string          `json:"errorMessage,omitempty"`
	ErrorDetails string          `json:"errorDetails,omitempty"`
	LogID        int             `json:"logID,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("message: %s, details: %s, logiD: %d, result: %s", a.ErrorMessage, a.ErrorDetails, a.LogID, a.Result)
}

type APIResponse[T any] struct {
	Result T   `json:"result,omitempty"`
	LogID  int `json:"logID,omitempty"`
}

type DNSInfo struct {
	DomainAdvancedDNSConfigID int            `json:"domainAdvancedDNSConfigID,omitempty"`
	DomainID                  int            `json:"domainID,omitempty"`
	DtCreated                 time.Time      `json:"dtCreated,omitzero"`
	DtModified                time.Time      `json:"dtModified,omitzero"`
	TimeToLive                int            `json:"timeToLive,omitempty"`
	SoAorigin                 string         `json:"soAorigin,omitempty"`
	SoArefresh                int            `json:"soArefresh,omitempty"`
	SoAretry                  int            `json:"soAretry,omitempty"`
	SoAexpire                 int            `json:"soAexpire,omitempty"`
	SoAnegcache               int            `json:"soAnegcache,omitempty"`
	ForwardingURL             string         `json:"forwardingURL,omitempty"`
	Gripping                  bool           `json:"gripping,omitempty"`
	Name                      string         `json:"name,omitempty"`
	DtSubmitted               time.Time      `json:"dtSubmitted,omitzero"`
	DtRequestedDNSChange      time.Time      `json:"dtRequestedDNSChange,omitzero"`
	Type                      string         `json:"type,omitempty"`
	UserManaged               bool           `json:"userManaged,omitempty"`
	EffectiveMgmtOption       string         `json:"effectiveMgmtOption,omitempty"`
	URLForwardRootOnly        bool           `json:"urlForwardRootOnly,omitempty"`
	EnableDNSSEC              bool           `json:"enableDNSSEC,omitempty"`
	DNSRecordSets             []DNSRecordSet `json:"dnsRecordSets,omitempty"`
}

type DNSRecordSet struct {
	Hostname string   `json:"hostname"`
	Type     string   `json:"type"`
	Records  []string `json:"records"`
}
