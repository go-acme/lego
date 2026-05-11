package internal

import (
	"fmt"
	"strings"
	"time"
)

type BaseResponse[E any, M any, D any] struct {
	Errors     []E `json:"errors"`
	Messages   []M `json:"messages"`
	Data       D   `json:"data"`
	StatusCode int `json:"status_code"`
}

type BasePaginatedResponse[E any, M any, D any] struct {
	BaseResponse[E, M, D]

	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Total       int    `json:"total"`
	Count       int    `json:"count"`
	PerPage     int    `json:"per_page"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	First       string `json:"first"`
	Last        string `json:"last"`
	Prev        string `json:"prev"`
	Next        string `json:"next"`
}

type Pager struct {
	Limit int    `url:"limit,omitempty"`
	Sort  string `url:"sort,omitempty"`
	Page  int    `url:"page,omitempty"`
}

type APIError BaseResponse[ErrorItem, string, []string]

func (a *APIError) Error() string {
	msg := new(strings.Builder)

	_, _ = fmt.Fprintf(msg, "%d: ", a.StatusCode)

	for _, item := range a.Errors {
		_, _ = fmt.Fprintf(msg, "%s: %s", item.Code, item.Message)

		if len(item.Details) > 0 {
			_, _ = fmt.Fprintf(msg, ": (%s)", strings.Join(item.Details, ", "))
		}
	}

	return msg.String()
}

type ErrorItem struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}

type APILoginRequest struct {
	Username string `url:"username"`
	Password string `url:"password"`
	APIKey   string `url:"api_key"`
}

type LoginResponse BaseResponse[string, Message, *TokenInfo]

type Message struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TokenInfo struct {
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	DNSDashboardLink int    `json:"dnsdashboard_link"`
}

type DomainsResponse struct {
	BasePaginatedResponse[string, string, []Domain]

	NameResults []NameResult `json:"name_results"`
	TermResults []TermResult `json:"term_results"`
}

type Domain struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	IDNName          string        `json:"idn_name"`
	TLD              string        `json:"tld"`
	CountryCode      string        `json:"country_code"`
	RegistryLock     *RegistryLock `json:"registry_lock"`
	Flagged          bool          `json:"flagged"`
	ActiveZone       *ActiveZone   `json:"active_zone"`
	ExternalComments string        `json:"external_comments"`
	Tags             []Tag         `json:"tags"`
}

type ActiveZone struct {
	ID                  string         `json:"id"`
	DefaultRecordTTL    int            `json:"default_record_ttl"`
	Signed              bool           `json:"signed"`
	ResourceRecordCount int            `json:"resource_record_count"`
	Secondary           *SecondaryZone `json:"secondary"`
	Networks            []string       `json:"networks"`
}

type SecondaryZone struct {
	Enabled          int      `json:"enabled"`
	PrimaryIP        string   `json:"primary_ip"`
	OtherIPs         []string `json:"other_ips"`
	DefaultRecordTTL int      `json:"default_record_ttl"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type RegistryLock struct {
	Enabled   bool      `json:"enabled"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TermResult struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type NameResult struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type RecordResponse BaseResponse[string, Message, RecordID]

type RecordID struct {
	ID string `json:"id"`
}

type RecordRequest struct {
	Name    string `url:"name,omitempty"`
	Type    string `url:"type,omitempty"`
	TTL     int    `url:"ttl,omitempty"`
	Content string `url:"value,omitempty"`
}
