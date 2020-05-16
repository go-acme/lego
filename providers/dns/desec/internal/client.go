package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

const baseURL = "https://desec.io/api/v1/"

// Client deSec API client.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string

	token string
}

// NewClient creates a new Client.
func NewClient(token string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    baseURL,
		token:      token,
	}
}

// GetTxtRRSet gets a RRSet.
// https://desec.readthedocs.io/en/latest/dns/rrsets.html#retrieving-a-specific-rrset
func (c *Client) GetTxtRRSet(domainName string, subName string) (*RRSet, error) {
	if subName == "" {
		subName = "@"
	}

	endpoint, err := c.createEndpoint("domains", domainName, "rrsets", subName, "TXT")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		var notFound NotFound
		err = json.Unmarshal(body, &notFound)
		if err != nil {
			return nil, fmt.Errorf("error: %d: %s", resp.StatusCode, string(body))
		}

		return nil, &notFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %d: %s", resp.StatusCode, string(body))
	}

	var rrSet RRSet
	err = json.Unmarshal(body, &rrSet)
	if err != nil {
		return nil, fmt.Errorf("failed to umarshal response body: %w", err)
	}

	return &rrSet, nil
}

// AddTxtRRSet creates a new RRSet.
// https://desec.readthedocs.io/en/latest/dns/rrsets.html#creating-a-tlsa-rrset
func (c *Client) AddTxtRRSet(rrSet RRSet) (*RRSet, error) {
	endpoint, err := c.createEndpoint("domains", rrSet.Domain, "rrsets")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	raw, err := json.Marshal(rrSet)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error: %d: %s", resp.StatusCode, string(body))
	}

	var newRRSet RRSet
	err = json.Unmarshal(body, &newRRSet)
	if err != nil {
		return nil, fmt.Errorf("failed to umarshal response body: %w", err)
	}

	return &newRRSet, nil
}

// UpdateTxtRRSet updates RRSet records.
// https://desec.readthedocs.io/en/latest/dns/rrsets.html#modifying-an-rrset
func (c *Client) UpdateTxtRRSet(domainName string, subName string, records []string) (*RRSet, error) {
	if subName == "" {
		subName = "@"
	}

	endpoint, err := c.createEndpoint("domains", domainName, "rrsets", subName, "TXT")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	raw, err := json.Marshal(RRSet{Records: records})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}

	// when a RRSet is deleted (empty records)
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %d: %s", resp.StatusCode, string(body))
	}

	var updatedRRSet RRSet
	err = json.Unmarshal(body, &updatedRRSet)
	if err != nil {
		return nil, fmt.Errorf("failed to umarshal response body: %w", err)
	}

	return &updatedRRSet, nil
}

// DeleteTxtRRSet deletes a RRset.
// https://desec.readthedocs.io/en/latest/dns/rrsets.html#deleting-an-rrset
func (c *Client) DeleteTxtRRSet(domainName string, subName string) error {
	if subName == "" {
		subName = "@"
	}

	endpoint, err := c.createEndpoint("domains", domainName, "rrsets", subName, "TXT")
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error: %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) createEndpoint(parts ...string) (string, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	endpoint, err := base.Parse(path.Join(base.Path, path.Join(parts...)))
	if err != nil {
		return "", err
	}

	endpoint.Path += "/"

	return endpoint.String(), nil
}
