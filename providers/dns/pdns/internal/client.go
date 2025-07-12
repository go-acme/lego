package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/miekg/dns"
)

// APIKeyHeader API key header.
const APIKeyHeader = "X-Api-Key"

// Client the PowerDNS API client.
type Client struct {
	serverName string
	apiKey     string

	apiVersion int

	Host       *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(host *url.URL, serverName string, apiVersion int, apiKey string) *Client {
	return &Client{
		serverName: serverName,
		apiKey:     apiKey,
		apiVersion: apiVersion,
		Host:       host,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) APIVersion() int {
	return c.apiVersion
}

func (c *Client) SetAPIVersion(ctx context.Context) error {
	var err error

	c.apiVersion, err = c.getAPIVersion(ctx)

	return err
}

func (c *Client) getAPIVersion(ctx context.Context) (int, error) {
	endpoint := c.joinPath("/", "api")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	result, err := c.do(req)
	if err != nil {
		return 0, err
	}

	var versions []apiVersion
	err = json.Unmarshal(result, &versions)
	if err != nil {
		return 0, err
	}

	latestVersion := 0
	for _, v := range versions {
		if v.Version > latestVersion {
			latestVersion = v.Version
		}
	}

	return latestVersion, err
}

func (c *Client) GetHostedZone(ctx context.Context, authZone string) (*HostedZone, error) {
	endpoint := c.joinPath("/", "servers", c.serverName, "zones", dns.Fqdn(authZone))

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var zone HostedZone
	err = json.Unmarshal(result, &zone)
	if err != nil {
		return nil, err
	}

	// convert pre-v1 API result
	if len(zone.Records) > 0 {
		zone.RRSets = []RRSet{}
		for _, record := range zone.Records {
			set := RRSet{
				Name:    record.Name,
				Type:    record.Type,
				Records: []Record{record},
			}
			zone.RRSets = append(zone.RRSets, set)
		}
	}

	return &zone, nil
}

func (c *Client) UpdateRecords(ctx context.Context, zone *HostedZone, sets RRSets) error {
	endpoint := c.joinPath("/", "servers", c.serverName, "zones", zone.ID)

	req, err := newJSONRequest(ctx, http.MethodPatch, endpoint, sets)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Notify(ctx context.Context, zone *HostedZone) error {
	if c.apiVersion < 1 || zone.Kind != "Master" && zone.Kind != "Slave" {
		return nil
	}

	endpoint := c.joinPath("/", "servers", c.serverName, "zones", zone.ID, "notify")

	req, err := newJSONRequest(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) joinPath(elem ...string) *url.URL {
	p := path.Join(elem...)

	if p != "/api" && c.apiVersion > 0 && !strings.HasPrefix(p, "/api/v") {
		p = path.Join("/api", "v"+strconv.Itoa(c.apiVersion), p)
	}

	return c.Host.JoinPath(p)
}

func (c *Client) do(req *http.Request) (json.RawMessage, error) {
	req.Header.Set(APIKeyHeader, c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnprocessableEntity && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	var msg json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// empty body
			return nil, nil
		}
		// other error
		return nil, err
	}

	// check for PowerDNS error message
	if len(msg) > 0 && msg[0] == '{' {
		var errInfo apiError
		err = json.Unmarshal(msg, &errInfo)
		if err != nil {
			return nil, errutils.NewUnmarshalError(req, resp.StatusCode, msg, err)
		}
		if errInfo.ShortMsg != "" {
			return nil, fmt.Errorf("error talking to PDNS API: %w", errInfo)
		}
	}

	return msg, nil
}

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(endpoint.String(), "/"), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// PowerDNS doesn't follow HTTP convention about the "Content-Type" header.
	if method != http.MethodGet && method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
