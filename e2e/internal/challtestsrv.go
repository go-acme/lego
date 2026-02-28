package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const persistLabel = "_validation-persist."

type ChallTestSrvClient struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewChallTestSrvClient() *ChallTestSrvClient {
	baseURL, _ := url.Parse("http://localhost:8055")

	return &ChallTestSrvClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *ChallTestSrvClient) SetPersistRecord(host, value string) error {
	return c.SetTXTRecord(persistLabel+strings.TrimPrefix(host, persistLabel), value)
}

func (c *ChallTestSrvClient) ClearPersistRecord(host string) error {
	return c.ClearTXTRecord(persistLabel + strings.TrimPrefix(host, persistLabel))
}

// SetTXTRecord to add a TXT record for a host.
// https://github.com/letsencrypt/pebble/tree/main/cmd/pebble-challtestsrv#dns-01
func (c *ChallTestSrvClient) SetTXTRecord(host, value string) error {
	endpoint := c.baseURL.JoinPath("set-txt")

	payload := map[string]string{
		"host":  host,
		"value": value,
	}

	return c.post(endpoint, payload)
}

// ClearTXTRecord to clear all TXT records for a host.
// https://github.com/letsencrypt/pebble/tree/main/cmd/pebble-challtestsrv#dns-01
func (c *ChallTestSrvClient) ClearTXTRecord(host string) error {
	endpoint := c.baseURL.JoinPath("clear-txt")

	payload := map[string]string{
		"host": host,
	}

	return c.post(endpoint, payload)
}

func (c *ChallTestSrvClient) post(endpoint *url.URL, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(endpoint.String(), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
