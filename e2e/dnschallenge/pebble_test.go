package dnschallenge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/challenge/dnspersist01"
	"github.com/stretchr/testify/require"
)

const persistLabel = "_validation-persist."

func updateDNS(t *testing.T, accountURI, issuerDomainName string) {
	t.Helper()

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "example.net", // Note: unused inside the tests.
		},
		Wildcard: true,
	}

	info, err := dnspersist01.GetChallengeInfo(authz, testPersistIssuer, accountURI, time.Time{})
	require.NoError(t, err)

	client := newChallTestSrvClient()

	err = client.SetPersistRecord(issuerDomainName, info.Value)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = client.ClearPersistRecord(issuerDomainName)
		require.NoError(t, err)
	})
}

type challTestSrvClient struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func newChallTestSrvClient() *challTestSrvClient {
	baseURL, _ := url.Parse("http://localhost:8055")

	return &challTestSrvClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *challTestSrvClient) SetPersistRecord(host, value string) error {
	return c.SetTXTRecord(persistLabel+strings.TrimPrefix(host, persistLabel), value)
}

func (c *challTestSrvClient) ClearPersistRecord(host string) error {
	return c.ClearTXTRecord(persistLabel + strings.TrimPrefix(host, persistLabel))
}

func (c *challTestSrvClient) SetTXTRecord(host, value string) error {
	endpoint := c.baseURL.JoinPath("set-txt")

	payload := map[string]string{
		"host":  host,
		"value": value,
	}

	return c.post(endpoint, payload)
}

func (c *challTestSrvClient) ClearTXTRecord(host string) error {
	endpoint := c.baseURL.JoinPath("clear-txt")

	payload := map[string]string{
		"host": host,
	}

	return c.post(endpoint, payload)
}

func (c *challTestSrvClient) post(endpoint *url.URL, payload any) error {
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
