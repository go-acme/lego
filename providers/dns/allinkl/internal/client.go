package internal

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

const (
	authEndpoint = "https://kasapi.kasserver.com/soap/KasAuth.php"
	apiEndpoint  = "https://kasapi.kasserver.com/soap/KasApi.php"
)

// Client a KAS server client.
type Client struct {
	login    string
	password string

	authEndpoint string
	apiEndpoint  string
	HTTPClient   *http.Client
	floodTime    time.Time
}

// NewClient creates a new Client.
func NewClient(login string, password string) *Client {
	return &Client{
		login:        login,
		password:     password,
		authEndpoint: authEndpoint,
		apiEndpoint:  apiEndpoint,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Authentication Creates a credential token.
// - sessionLifetime: Validity of the token in seconds.
// - sessionUpdateLifetime: with `true` the session is extended with every request.
func (c Client) Authentication(sessionLifetime int, sessionUpdateLifetime bool) (string, error) {
	hash := sha1.New()
	hash.Write([]byte(c.password))

	sul := "N"
	if sessionUpdateLifetime {
		sul = "Y"
	}

	ar := AuthRequest{
		Login:                 c.login,
		AuthData:              fmt.Sprintf("%x", hash.Sum(nil)),
		AuthType:              "sha1",
		SessionLifetime:       sessionLifetime,
		SessionUpdateLifetime: sul,
	}

	body, err := json.Marshal(ar)
	if err != nil {
		return "", fmt.Errorf("request marshal: %w", err)
	}

	payload := []byte(strings.TrimSpace(fmt.Sprintf(kasAuthEnvelope, body)))

	req, err := http.NewRequest(http.MethodPost, c.authEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("request creation: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request execution: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("invalid status code: %d %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response read: %w", err)
	}

	var e KasAuthEnvelope
	decoder := xml.NewTokenDecoder(Trimmer{decoder: xml.NewDecoder(bytes.NewReader(data))})
	err = decoder.Decode(&e)
	if err != nil {
		return "", fmt.Errorf("response xml decode: %w", err)
	}

	if e.Body.Fault != nil {
		return "", e.Body.Fault
	}

	return e.Body.KasAuthResponse.Return.Text, nil
}

// GetDNSSettings Reading out the DNS settings of a zone.
// - zone: host zone.
// - recordID: the ID of the resource record (optional).
func (c *Client) GetDNSSettings(credentialToken, zone, recordID string) ([]ReturnInfo, error) {
	requestParams := map[string]string{"zone_host": zone}

	if recordID != "" {
		requestParams["record_id"] = recordID
	}

	item, err := c.do(credentialToken, "get_dns_settings", requestParams)
	if err != nil {
		return nil, err
	}

	raw := getValue(item)

	var g GetDNSSettingsAPIResponse
	err = mapstructure.Decode(raw, &g)
	if err != nil {
		return nil, fmt.Errorf("response struct decode: %w", err)
	}

	c.updateFloodTime(g.Response.KasFloodDelay)

	return g.Response.ReturnInfo, nil
}

// AddDNSSettings Creation of a DNS resource record.
func (c *Client) AddDNSSettings(credentialToken string, record DNSRequest) (string, error) {
	item, err := c.do(credentialToken, "add_dns_settings", record)
	if err != nil {
		return "", err
	}

	raw := getValue(item)

	var g AddDNSSettingsAPIResponse
	err = mapstructure.Decode(raw, &g)
	if err != nil {
		return "", fmt.Errorf("response struct decode: %w", err)
	}

	c.updateFloodTime(g.Response.KasFloodDelay)

	return g.Response.ReturnInfo, nil
}

// DeleteDNSSettings Deleting a DNS Resource Record.
func (c *Client) DeleteDNSSettings(credentialToken, recordID string) (bool, error) {
	requestParams := map[string]string{"record_id": recordID}

	item, err := c.do(credentialToken, "delete_dns_settings", requestParams)
	if err != nil {
		return false, err
	}

	raw := getValue(item)

	var g DeleteDNSSettingsAPIResponse
	err = mapstructure.Decode(raw, &g)
	if err != nil {
		return false, fmt.Errorf("response struct decode: %w", err)
	}

	c.updateFloodTime(g.Response.KasFloodDelay)

	return g.Response.ReturnInfo, nil
}

func (c Client) do(credentialToken, action string, requestParams interface{}) (*Item, error) {
	time.Sleep(time.Until(c.floodTime))

	ar := KasRequest{
		Login:         c.login,
		AuthType:      "session",
		AuthData:      credentialToken,
		Action:        action,
		RequestParams: requestParams,
	}

	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("request marshal: %w", err)
	}

	payload := []byte(strings.TrimSpace(fmt.Sprintf(kasAPIEnvelope, body)))

	req, err := http.NewRequest(http.MethodPost, c.apiEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("request creation: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid status code: %d %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response read: %w", err)
	}

	var e KasAPIResponseEnvelope
	decoder := xml.NewTokenDecoder(Trimmer{decoder: xml.NewDecoder(bytes.NewReader(data))})
	err = decoder.Decode(&e)
	if err != nil {
		return nil, fmt.Errorf("response xml decode: %w", err)
	}

	if e.Body.Fault != nil {
		return nil, e.Body.Fault
	}

	return e.Body.KasAPIResponse.Return, nil
}

func (c *Client) updateFloodTime(delay float64) {
	c.floodTime = time.Now().Add(time.Duration(delay * float64(time.Second)))
}

func getValue(item *Item) interface{} {
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
		var v []interface{}
		for _, i := range item.Items {
			v = append(v, getValue(i))
		}

		return v

	case len(item.Items) > 0:
		v := map[string]interface{}{}
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
