package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	apiBaseURL = "https://admin.vshosting.cloud/clouddns"
	loginURL   = "https://admin.vshosting.cloud/api/public/auth/login"
)

// Client handles all communication with CloudDNS API.
type Client struct {
	AccessToken string
	ClientID    string
	Email       string
	Password    string
	TTL         int
	HTTPClient  *http.Client

	apiBaseURL string
	loginURL   string
}

// NewClient returns a Client instance configured to handle CloudDNS API communication.
func NewClient(clientID, email, password string, ttl int) *Client {
	return &Client{
		ClientID:   clientID,
		Email:      email,
		Password:   password,
		TTL:        ttl,
		HTTPClient: &http.Client{},
		apiBaseURL: apiBaseURL,
		loginURL:   loginURL,
	}
}

// AddRecord is a high level method to add a new record into CloudDNS zone.
func (c *Client) AddRecord(zone, recordName, recordValue string) error {
	domain, err := c.getDomain(zone)
	if err != nil {
		return err
	}

	record := Record{DomainID: domain.ID, Name: recordName, Value: recordValue, Type: "TXT"}

	err = c.addTxtRecord(record)
	if err != nil {
		return err
	}

	return c.publishRecords(domain.ID)
}

// DeleteRecord is a high level method to remove a record from zone.
func (c *Client) DeleteRecord(zone, recordName string) error {
	domain, err := c.getDomain(zone)
	if err != nil {
		return err
	}

	record, err := c.getRecord(domain.ID, recordName)
	if err != nil {
		return err
	}

	err = c.deleteRecord(record)
	if err != nil {
		return err
	}

	return c.publishRecords(domain.ID)
}

func (c *Client) addTxtRecord(record Record) error {
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, err = c.doAPIRequest(http.MethodPost, "record-txt", bytes.NewReader(body))
	return err
}

func (c *Client) deleteRecord(record Record) error {
	endpoint := fmt.Sprintf("record/%s", record.ID)
	_, err := c.doAPIRequest(http.MethodDelete, endpoint, nil)
	return err
}

func (c *Client) getDomain(zone string) (Domain, error) {
	searchQuery := SearchQuery{
		Search: []Search{
			{Name: "clientId", Operator: "eq", Value: c.ClientID},
			{Name: "domainName", Operator: "eq", Value: zone},
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return Domain{}, err
	}

	resp, err := c.doAPIRequest(http.MethodPost, "domain/search", bytes.NewReader(body))
	if err != nil {
		return Domain{}, err
	}

	var result SearchResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Domain{}, err
	}

	if len(result.Items) == 0 {
		return Domain{}, fmt.Errorf("domain not found: %s", zone)
	}

	return result.Items[0], nil
}

func (c *Client) getRecord(domainID, recordName string) (Record, error) {
	endpoint := fmt.Sprintf("domain/%s", domainID)
	resp, err := c.doAPIRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return Record{}, err
	}

	var result DomainInfo
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return Record{}, err
	}

	for _, record := range result.LastDomainRecordList {
		if record.Name == recordName && record.Type == "TXT" {
			return record, nil
		}
	}

	return Record{}, fmt.Errorf("record not found: domainID %s, name %s", domainID, recordName)
}

func (c *Client) publishRecords(domainID string) error {
	body, err := json.Marshal(DomainInfo{SoaTTL: c.TTL})
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("domain/%s/publish", domainID)
	_, err = c.doAPIRequest(http.MethodPut, endpoint, bytes.NewReader(body))
	return err
}

func (c *Client) login() error {
	authorization := Authorization{Email: c.Email, Password: c.Password}

	body, err := json.Marshal(authorization)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.loginURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	content, err := c.doRequest(req)
	if err != nil {
		return err
	}

	var result AuthResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		return err
	}

	c.AccessToken = result.Auth.AccessToken

	return nil
}

func (c *Client) doAPIRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	if c.AccessToken == "" {
		err := c.login()
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/%s", c.apiBaseURL, endpoint)

	req, err := c.newRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	content, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (c *Client) newRequest(method, reqURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	return req, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, readError(req, resp)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func readError(req *http.Request, resp *http.Response) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(toUnreadableBodyMessage(req, content))
	}

	var errInfo APIError
	err = json.Unmarshal(content, &errInfo)
	if err != nil {
		return fmt.Errorf("APIError unmarshaling error: %w: %s", err, toUnreadableBodyMessage(req, content))
	}

	return fmt.Errorf("HTTP %d: code %v: %s", resp.StatusCode, errInfo.Error.Code, errInfo.Error.Message)
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
