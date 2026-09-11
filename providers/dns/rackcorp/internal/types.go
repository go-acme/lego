package internal

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}

type APIResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type Record struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Status     string `json:"status,omitempty"`
	Type       string `json:"type,omitempty"`
	RegionID   int64  `json:"regionid,omitempty"`
	RegionName string `json:"regionname,omitempty"`
	Lookup     string `json:"lookup,omitempty"`
	Priority   int    `json:"priority,omitempty"`
	Weight     int    `json:"weight,omitempty"`
	Port       int    `json:"port,omitempty"`
	CAATag     int    `json:"caatag,omitempty"`
	CAAFlag    int    `json:"caaflag,omitempty"`
	CustomerID int64  `json:"customerid,omitempty"`
	StdName    string `json:"stdname,omitempty"`
	Serial     string `json:"serial,omitempty"`
	DomainID   int64  `json:"domainid,omitempty"`
	Data       string `json:"data,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
}

type Domain struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	CustomerID int64  `json:"customerid"`
	STDName    string `json:"stdname"`
	Serial     string `json:"serial"`
}
