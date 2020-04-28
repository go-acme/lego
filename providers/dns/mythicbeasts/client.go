package mythicbeasts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

const (
	apiBaseURL  = "https://api.mythic-beasts.com/beta/dns"
	authBaseURL = "https://auth.mythic-beasts.com/login"
)

type authResponse struct {
	// The bearer token for use in API requests
	Token string `json:"access_token"`

	// The maximum lifetime of the token in seconds
	Lifetime int `json:"expires_in"`

	// The token type (must be 'bearer')
	TokenType string `json:"token_type"`
}

type authResponseError struct {
	ErrorMsg         string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (a authResponseError) Error() string {
	return fmt.Sprintf("%s: %s", a.ErrorMsg, a.ErrorDescription)
}

type createTXTRequest struct {
	Records []createTXTRecord `json:"records"`
}

type createTXTRecord struct {
	Host string `json:"host"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
	Data string `json:"data"`
}

type createTXTResponse struct {
	Added   int    `json:"records_added"`
	Removed int    `json:"records_removed"`
	Message string `json:"message"`
}

type deleteTXTResponse struct {
	Removed int    `json:"records_removed"`
	Message string `json:"message"`
}

// Logs into mythic beasts and acquires a bearer token for use in future API calls.
// https://www.mythic-beasts.com/support/api/auth#sec-obtaining-a-token
func (d *DNSProvider) login() error {
	if d.token != "" {
		// Already authenticated, stop now
		return nil
	}

	reqBody := strings.NewReader("grant_type=client_credentials")

	req, err := http.NewRequest(http.MethodPost, d.config.AuthAPIEndpoint.String(), reqBody)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(d.config.UserName, d.config.Password)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode < 400 || resp.StatusCode > 499 {
			return fmt.Errorf("login: unknown error in auth API: %d", resp.StatusCode)
		}

		// Returned body should be a JSON thing
		errResp := &authResponseError{}
		err = json.Unmarshal(body, errResp)
		if err != nil {
			return fmt.Errorf("login: error parsing error: %w", err)
		}

		return fmt.Errorf("login: %d: %w", resp.StatusCode, errResp)
	}

	authResp := authResponse{}
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		return fmt.Errorf("login: error parsing response: %w", err)
	}

	if authResp.TokenType != "bearer" {
		return fmt.Errorf("login: received unexpected token type: %s", authResp.TokenType)
	}

	d.token = authResp.Token

	// Success
	return nil
}

// https://www.mythic-beasts.com/support/api/dnsv2#ep-get-zoneszonerecords
func (d *DNSProvider) createTXTRecord(zone string, leaf string, value string) error {
	if d.token == "" {
		return fmt.Errorf("createTXTRecord: not logged in")
	}

	createReq := createTXTRequest{
		Records: []createTXTRecord{{
			Host: leaf,
			TTL:  d.config.TTL,
			Type: "TXT",
			Data: value,
		}},
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		return fmt.Errorf("createTXTRecord: marshaling request body failed: %w", err)
	}

	endpoint, err := d.config.APIEndpoint.Parse(path.Join(d.config.APIEndpoint.Path, "zones", zone, "records", leaf, "TXT"))
	if err != nil {
		return fmt.Errorf("createTXTRecord: failed to parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("createTXTRecord: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("createTXTRecord: unable to perform HTTP request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("createTXTRecord: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("createTXTRecord: error in API: %d", resp.StatusCode)
	}

	createResp := createTXTResponse{}
	err = json.Unmarshal(body, &createResp)
	if err != nil {
		return fmt.Errorf("createTXTRecord: error parsing response: %w", err)
	}

	if createResp.Added != 1 {
		return errors.New("createTXTRecord: did not add TXT record for some reason")
	}

	// Success
	return nil
}

// https://www.mythic-beasts.com/support/api/dnsv2#ep-delete-zoneszonerecords
func (d *DNSProvider) removeTXTRecord(zone string, leaf string, value string) error {
	if d.token == "" {
		return fmt.Errorf("removeTXTRecord: not logged in")
	}

	endpoint, err := d.config.APIEndpoint.Parse(path.Join(d.config.APIEndpoint.Path, "zones", zone, "records", leaf, "TXT"))
	if err != nil {
		return fmt.Errorf("createTXTRecord: failed to parse URL: %w", err)
	}

	query := endpoint.Query()
	query.Add("data", value)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("removeTXTRecord: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("removeTXTRecord: unable to perform HTTP request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("removeTXTRecord: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("removeTXTRecord: error in API: %d", resp.StatusCode)
	}

	deleteResp := deleteTXTResponse{}
	err = json.Unmarshal(body, &deleteResp)
	if err != nil {
		return fmt.Errorf("removeTXTRecord: error parsing response: %w", err)
	}

	if deleteResp.Removed != 1 {
		return errors.New("deleteTXTRecord: did not add TXT record for some reason")
	}

	// Success
	return nil
}
