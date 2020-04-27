package mythicbeasts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
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
	// The error
	Error string `json:"error"`
	// A description of the error
	ErrorDescription string `json:"error_description"`
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

// Logs into mythic beasts and acquires a bearer token for use in future
// API calls
func (d *DNSProvider) login() error {
	if d.token != "" {
		// Already authenticated, stop now
		return nil
	}
	sendbody := strings.NewReader("grant_type=client_credentials")

	req, err := http.NewRequest("POST", d.config.AuthAPIEndpoint.String(), sendbody)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(d.config.UserName, d.config.Password)

	resp, err := d.config.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return fmt.Errorf("login: %w", readErr)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
			// Returned body should be a JSON thing
			err := authResponseError{}
			jsonErr := json.Unmarshal(body, &err)
			if jsonErr != nil {
				return fmt.Errorf("login: Error parsing error: %w", jsonErr)
			}
			return fmt.Errorf("login: %d: %s: %s", resp.StatusCode, err.Error, err.ErrorDescription)
		}
		return fmt.Errorf("login: Unknown error in auth API: %d", resp.StatusCode)
	}

	authresp := authResponse{}
	jsonErr := json.Unmarshal(body, &authresp)
	if jsonErr != nil {
		return fmt.Errorf("login: Error parsing response: %w", jsonErr)
	}

	if authresp.TokenType != "bearer" {
		return fmt.Errorf("login: Received unexpected token type: %s", authresp.TokenType)
	}

	d.token = authresp.Token
	return nil // Success
}

func (d *DNSProvider) createTXTRecord(zone string, leaf string, value string) error {
	if d.token == "" {
		return fmt.Errorf("createTXTRecord: Not logged in")
	}

	createbody := createTXTRequest{
		Records: []createTXTRecord{
			{
				Host: leaf,
				TTL:  d.config.TTL,
				Type: "TXT",
				Data: value,
			},
		},
	}

	sendbody, err := json.Marshal(createbody)
	if err != nil {
		return fmt.Errorf("createTXTRecord: Marshaling request body failed: %w", err)
	}

	req, err := http.NewRequest("POST", d.txtURL(zone, leaf), bytes.NewReader(sendbody))

	if err != nil {
		return fmt.Errorf("createTXTRecord: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("createTXTRecord: Unable to perform HTTP request: %w", err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return fmt.Errorf("createTXTRecord: %w", readErr)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("createTXTRecord: Error in API: %d", resp.StatusCode)
	}

	createresp := createTXTResponse{}
	jsonErr := json.Unmarshal(body, &createresp)
	if jsonErr != nil {
		return fmt.Errorf("createTXTRecord: Error parsing response: %w", jsonErr)
	}

	if createresp.Added != 1 {
		return fmt.Errorf("mythicbeasts: Did not add TXT record for some reason")
	}

	return nil // Success
}

func (d *DNSProvider) removeTXTRecord(zone string, leaf string, value string) error {
	if d.token == "" {
		return fmt.Errorf("removeTXTRecord: Not logged in")
	}

	req, err := http.NewRequest("DELETE", d.txtURL(zone, leaf)+"?data="+value, nil)

	if err != nil {
		return fmt.Errorf("removeTXTRecord: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("removeTXTRecord: Unable to perform HTTP request: %w", err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return fmt.Errorf("removeTXTRecord: %w", readErr)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("removeTXTRecord: Error in API: %d", resp.StatusCode)
	}

	deleteresp := deleteTXTResponse{}
	jsonErr := json.Unmarshal(body, &deleteresp)
	if jsonErr != nil {
		return fmt.Errorf("removeTXTRecord: Error parsing response: %w", jsonErr)
	}

	if deleteresp.Removed != 1 {
		return fmt.Errorf("mythicbeasts: deleteTXTRecord: Did not add TXT record for some reason")
	}

	return nil // Success
}

// Internal function to determine the full URL for a given zone+leaf
func (d *DNSProvider) txtURL(zone string, leaf string) string {
	u := *d.config.APIEndpoint
	u.Path = path.Join(u.Path, "zones", zone, "records", leaf, "TXT")
	return u.String()
}
