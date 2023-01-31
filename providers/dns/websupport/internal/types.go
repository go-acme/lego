package internal

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.Code, a.Message)
}

type Record struct {
	ID      int    `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"` // subdomain name or @ if you don't want subdomain
	Content string `json:"content,omitempty"`
	TTL     int    `json:"ttl,omitempty"` // default 600
	Zone    *Zone  `json:"zone"`
}

type Zone struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	UpdateTime int    `json:"updateTime"`
}

type Response struct {
	Status string          `json:"status"`
	Item   *Record         `json:"item"`
	Errors json.RawMessage `json:"errors"`
}

type ListResponse struct {
	Items []Record `json:"items"`
	Pager Pager    `json:"pager"`
}

type Pager struct {
	Page     int `json:"page"`
	PageSize int `json:"pagesize"`
	Items    int `json:"items"`
}

type Errors struct {
	Name    []string `json:"name"`
	Content []string `json:"content"`
}

func (e *Errors) Error() string {
	var msg string
	for i, s := range e.Name {
		msg += s
		if i != len(e.Name)-1 {
			msg += ": "
		}
	}

	for i, s := range e.Content {
		msg += s
		if i != len(e.Content)-1 {
			msg += ": "
		}
	}

	return msg
}

// ParseError extract error from Response.
func ParseError(resp *Response) error {
	var apiError Errors
	err := json.Unmarshal(resp.Errors, &apiError)
	if err != nil {
		return err
	}

	return &apiError
}

type User struct {
	ID                      int       `json:"id"`
	Login                   string    `json:"login"`
	ParentID                int       `json:"parentId"`
	Active                  bool      `json:"active"`
	CreateTime              int       `json:"createTime"`
	Group                   string    `json:"group"`
	Email                   string    `json:"email"`
	Phone                   string    `json:"phone"`
	ContactPerson           string    `json:"contactPerson"`
	AwaitingTosConfirmation string    `json:"awaitingTosConfirmation"`
	UserLanguage            string    `json:"userLanguage"`
	Credit                  int       `json:"credit"`
	VerifyURL               string    `json:"verifyUrl"`
	Billing                 []Billing `json:"billing"`
	Market                  Market    `json:"market"`
}

type Billing struct {
	ID           int    `json:"id"`
	Profile      string `json:"profile"`
	IsDefault    bool   `json:"isDefault"`
	Name         string `json:"name"`
	City         string `json:"city"`
	Street       string `json:"street"`
	CompanyRegID int    `json:"companyRegId"`
	TaxID        int    `json:"taxId"`
	VatID        int    `json:"vatId"`
	Zip          string `json:"zip"`
	Country      string `json:"country"`
	ISIC         string `json:"isic"`
}

type Market struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Currency   string `json:"currency"`
}
