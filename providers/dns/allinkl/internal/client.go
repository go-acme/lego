package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/mitchellh/mapstructure"
)

const apiEndpoint = "https://kasapi.kasserver.com/soap/KasApi.php"

type Authentication interface {
	Authentication(ctx context.Context, sessionLifetime int, sessionUpdateLifetime bool) (string, error)
}

// Client a KAS server client.
type Client struct {
	login string

	floodTime   time.Time
	muFloodTime sync.Mutex

	baseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(login string) *Client {
	return &Client{
		login:      login,
		baseURL:    apiEndpoint,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetDNSSettings Reading out the DNS settings of a zone.
// - zone: host zone.
// - recordID: the ID of the resource record (optional).
func (c *Client) GetDNSSettings(ctx context.Context, zone, recordID string) ([]ReturnInfo, error) {
	requestParams := map[string]string{"zone_host": zone}

	if recordID != "" {
		requestParams["record_id"] = recordID
	}

	req, err := c.newRequest(ctx, "get_dns_settings", requestParams)
	if err != nil {
		return nil, err
	}

	var g GetDNSSettingsAPIResponse
	err = c.do(req, &g)
	if err != nil {
		return nil, err
	}

	c.updateFloodTime(g.Response.KasFloodDelay)

	return g.Response.ReturnInfo, nil
}

// AddDNSSettings Creation of a DNS resource record.
func (c *Client) AddDNSSettings(ctx context.Context, record DNSRequest) (string, error) {
	req, err := c.newRequest(ctx, "add_dns_settings", record)
	if err != nil {
		return "", err
	}

	var g AddDNSSettingsAPIResponse
	err = c.do(req, &g)
	if err != nil {
		return "", err
	}

	c.updateFloodTime(g.Response.KasFloodDelay)

	return g.Response.ReturnInfo, nil
}

// DeleteDNSSettings Deleting a DNS Resource Record.
func (c *Client) DeleteDNSSettings(ctx context.Context, recordID string) (bool, error) {
	requestParams := map[string]string{"record_id": recordID}

	req, err := c.newRequest(ctx, "delete_dns_settings", requestParams)
	if err != nil {
		return false, err
	}

	var g DeleteDNSSettingsAPIResponse
	err = c.do(req, &g)
	if err != nil {
		return false, err
	}

	c.updateFloodTime(g.Response.KasFloodDelay)

	return g.Response.ReturnInfo, nil
}

func (c *Client) newRequest(ctx context.Context, action string, requestParams any) (*http.Request, error) {
	ar := KasRequest{
		Login:         c.login,
		AuthType:      "session",
		AuthData:      getToken(ctx),
		Action:        action,
		RequestParams: requestParams,
	}

	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	payload := []byte(strings.TrimSpace(fmt.Sprintf(kasAPIEnvelope, body)))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	return req, nil
}

func (c *Client) do(req *http.Request, result any) error {
	c.muFloodTime.Lock()
	time.Sleep(time.Until(c.floodTime))
	c.muFloodTime.Unlock()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	envlp, err := decodeXML[KasAPIResponseEnvelope](resp.Body)
	if err != nil {
		return err
	}

	if envlp.Body.Fault != nil {
		return envlp.Body.Fault
	}

	raw := getValue(envlp.Body.KasAPIResponse.Return)

	err = mapstructure.Decode(raw, result)
	if err != nil {
		return fmt.Errorf("response struct decode: %w", err)
	}

	return nil
}

func (c *Client) updateFloodTime(delay float64) {
	c.muFloodTime.Lock()
	c.floodTime = time.Now().Add(time.Duration(delay * float64(time.Second)))
	c.muFloodTime.Unlock()
}

func getValue(item *Item) any {
	switch {
	case item.Raw != "":
		v, _ := strconv.ParseBool(item.Raw)
		return v

	case item.Text != "":
		switch item.Type {
		case "xsd:string":
			return item.Text
		case "xsd:float":
			v, _ := strconv.ParseFloat(item.Text, 64)
			return v
		case "xsd:int":
			v, _ := strconv.ParseInt(item.Text, 10, 64)
			return v
		default:
			return item.Text
		}

	case item.Value != nil:
		return getValue(item.Value)

	case len(item.Items) > 0 && item.Type == "SOAP-ENC:Array":
		var v []any
		for _, i := range item.Items {
			v = append(v, getValue(i))
		}

		return v

	case len(item.Items) > 0:
		v := map[string]any{}
		for _, i := range item.Items {
			v[getKey(i)] = getValue(i)
		}

		return v

	default:
		return ""
	}
}

func getKey(item *Item) string {
	if item.Key == nil {
		return ""
	}

	return item.Key.Text
}
