package internal

import (
	"fmt"
	"time"
)

type APIError struct {
	ErrorMessage string `json:"errorMessage,omitempty"`
	ErrorDetails string `json:"errorDetails,omitempty"`
	LogID        int    `json:"logID,omitempty"`
	Result       string `json:"result,omitempty"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("message: %s, details: %s, logiD: %d, result: %s", a.ErrorMessage, a.ErrorDetails, a.LogID, a.Result)
}

type APIResponse struct {
	ErrorMessage string  `json:"errorMessage,omitempty"`
	ErrorDetails string  `json:"errorDetails,omitempty"`
	LogID        int     `json:"logID,omitempty"`
	Result       *Result `json:"result,omitempty"`
}

type Result struct {
	DomainAdvancedDNSConfigID int            `json:"domainAdvancedDNSConfigID,omitempty"`
	DomainID                  int            `json:"domainID,omitempty"`
	DtCreated                 time.Time      `json:"dtCreated,omitempty"`
	DtModified                time.Time      `json:"dtModified,omitempty"`
	TimeToLive                int            `json:"timeToLive,omitempty"`
	SoAorigin                 string         `json:"soAorigin,omitempty"`
	SoArefresh                int            `json:"soArefresh,omitempty"`
	SoAretry                  int            `json:"soAretry,omitempty"`
	SoAexpire                 int            `json:"soAexpire,omitempty"`
	SoAnegcache               int            `json:"soAnegcache,omitempty"`
	ForwardingURL             string         `json:"forwardingURL,omitempty"`
	Gripping                  bool           `json:"gripping,omitempty"`
	Name                      string         `json:"name,omitempty"`
	DtSubmitted               time.Time      `json:"dtSubmitted,omitempty"`
	DtRequestedDNSChange      time.Time      `json:"dtRequestedDNSChange,omitempty"`
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
