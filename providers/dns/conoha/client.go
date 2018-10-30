package conoha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// IdentityRequest is an authentication request body.
type IdentityRequest struct {
	Auth Auth `json:"auth"`
}

// Auth is an authentication information.
type Auth struct {
	TenantID            string              `json:"tenantId"`
	PasswordCredentials PasswordCredentials `json:"passwordCredentials"`
}

// PasswordCredentials is API-user's credentials.
type PasswordCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// IdentityResponse is an authentication response body.
type IdentityResponse struct {
	Access Access `json:"access"`
}

// Access is an identity information.
type Access struct {
	Token Token `json:"token"`
}

// Token is an api access token.
type Token struct {
	ID string `json:"id"`
}

// DomainListResponse is a response of a domain listing request.
type DomainListResponse struct {
	Domains []Domain `json:"domains"`
}

// Domain is a hosted domain entry.
type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecordListResponse is a response of record listing request.
type RecordListResponse struct {
	Records []Record `json:"records"`
}

// Record is a record entry.
type Record struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

// Client is a ConoHa API client.
type Client struct {
	*http.Client
	token    string
	endpoint string
}

// NewClient returns a client instance logged into the ConoHa service.
func NewClient(region, tenant, username, password string) (*Client, error) {
	c := &Client{
		Client:   &http.Client{},
		endpoint: "https://identity." + region + ".conoha.io/v2.0",
	}

	req := &IdentityRequest{
		Auth{
			tenant,
			PasswordCredentials{username, password},
		},
	}
	resp := &IdentityResponse{}

	err := c.request("POST", "/tokens", req, resp)
	if err != nil {
		return nil, err
	}

	c.token = resp.Access.Token.ID
	c.endpoint = "https://dns-service." + region + ".conoha.io/v1"
	return c, nil
}

// GetDomainID returns an ID of specified domain.
func (c *Client) GetDomainID(domainName string) (string, error) {
	resp := &DomainListResponse{}
	err := c.request("GET", "/domains", nil, resp)
	if err != nil {
		return "", err
	}

	for _, domain := range resp.Domains {
		if domain.Name == domainName {
			return domain.ID, nil
		}
	}
	return "", errors.New("no such domain")
}

// GetRecordID returns an ID of specified record.
func (c *Client) GetRecordID(domainID, recordName, recordType, data string) (string, error) {
	resp := &RecordListResponse{}
	err := c.request("GET", "/domains/"+domainID+"/records", nil, resp)
	if err != nil {
		return "", err
	}

	for _, record := range resp.Records {
		if record.Name == recordName && record.Type == recordType && record.Data == data {
			return record.ID, nil
		}
	}
	return "", errors.New("no such record")
}

// CreateRecord adds new record.
func (c *Client) CreateRecord(domainID, recordName, recordType, data string, ttl int) error {
	req := &Record{"", recordName, recordType, data, ttl}
	return c.request("POST", "/v1/domains/"+domainID+"/records", req, nil)
}

// DeleteRecord removes specified record.
func (c *Client) DeleteRecord(domainID, recordID string) error {
	return c.request("DELETE", "/v1/domains/"+domainID+"/records/"+recordID, nil, nil)
}

func (c *Client) request(method, path string, payload, result interface{}) error {
	body := bytes.NewReader(nil)

	if payload != nil {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.endpoint+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", c.token)

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	if result != nil {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(respBody, result)
	}

	return nil
}
